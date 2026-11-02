package emailverifier

import (
	"context"
	"errors"
	"net"
	"net/textproto"
	"net/url"
	"strings"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/dns/dnsmessage"
	"golang.org/x/net/proxy"
)

func TestCheckSMTPUnSupportedVendor(t *testing.T) {
	err := verifier.EnableAPIVerifier("unsupported_vendor")
	assert.Error(t, err)
}

func TestCheckSMTPOK_HostExists(t *testing.T) {
	domain := "github.com"

	smtp, err := verifier.CheckSMTP(domain, "")
	expected := SMTP{
		HostExists: true,
		FullInbox:  false,
		CatchAll:   true,
		Disabled:   false,
	}
	require.NoError(t, err)
	assert.Equal(t, &expected, smtp)
}

func TestCheckSMTPOK_CatchAllHost(t *testing.T) {
	domain := "gmail.com"

	smtp, err := verifier.CheckSMTP(domain, "")
	expected := SMTP{
		HostExists: true,
		FullInbox:  false,
		CatchAll:   false,
		Disabled:   false,
	}
	require.NoError(t, err)
	assert.Equal(t, &expected, smtp)
}

func TestCheckSMTPOK_NoCatchAllHost(t *testing.T) {
	domain := "gmail.com"

	smtp, err := verifier.CheckSMTP(domain, "")
	expected := SMTP{
		HostExists: true,
		FullInbox:  false,
		CatchAll:   false,
		Disabled:   false,
	}
	require.NoError(t, err)
	assert.Equal(t, &expected, smtp)
}

func TestCheckSMTPOK_NoCatchAllHostCatchAllCheckDisabled(t *testing.T) {
	domain := "gmail.com"

	var verifier = NewVerifier().EnableSMTPCheck().DisableCatchAllCheck()
	smtp, err := verifier.CheckSMTP(domain, "")
	expected := SMTP{
		HostExists: true,
		FullInbox:  false,
		CatchAll:   true,
		Disabled:   false,
	}
	require.NoError(t, err)
	assert.Equal(t, &expected, smtp)
}

func TestCheckSMTPOK_UpdateFromEmail(t *testing.T) {
	domain := "github.com"
	verifier.FromEmail("from@email.top")

	smtp, err := verifier.CheckSMTP(domain, "")
	expected := SMTP{
		HostExists:  true,
		FullInbox:   false,
		CatchAll:    true,
		Deliverable: false,
		Disabled:    false,
	}
	require.NoError(t, err)
	assert.Equal(t, &expected, smtp)
}

func TestCheckSMTPOK_UpdateHelloName(t *testing.T) {
	domain := "github.com"
	verifier.HelloName("email.top")

	smtp, err := verifier.CheckSMTP(domain, "")
	expected := SMTP{
		HostExists:  true,
		FullInbox:   false,
		CatchAll:    true,
		Deliverable: false,
		Disabled:    false,
	}
	require.NoError(t, err)
	assert.Equal(t, &expected, smtp)
}

func TestCheckSMTPOK_WithNoExistUsername(t *testing.T) {
	domain := "github.com"
	username := "testing"

	smtp, err := verifier.CheckSMTP(domain, username)
	expected := SMTP{
		HostExists: true,
		FullInbox:  false,
		CatchAll:   true,
		Disabled:   false,
	}
	require.NoError(t, err)
	assert.Equal(t, &expected, smtp)
}

func TestCheckSMTP_DisabledSMTPCheck(t *testing.T) {
	domain := "github.com"

	verifier.DisableSMTPCheck()
	smtp, err := verifier.CheckSMTP(domain, "username")
	verifier.EnableSMTPCheck()

	require.NoError(t, err)
	assert.Nil(t, smtp)
}

func TestCheckSMTPOK_HostNotExists(t *testing.T) {
	domain := "notExistHost.com"

	smtp, err := verifier.CheckSMTP(domain, "")
	require.ErrorContains(t, err, ErrNoSuchHost)
	assert.Equal(t, &SMTP{}, smtp)
}

func TestNewSMTPClientOK(t *testing.T) {
	domain := "gmail.com"
	timeout := 5 * time.Second
	ret, _, err := newSMTPClient(domain, "", net.DefaultResolver, timeout, timeout)
	assert.NotNil(t, ret)
	assert.NoError(t, err)
}

func TestNewSMTPClientFailed_WithInvalidProxy(t *testing.T) {
	domain := "gmail.com"
	proxyURI := "socks5://user:password@127.0.0.1:1080?timeout=5s"
	timeout := 5 * time.Second
	ret, _, err := newSMTPClient(domain, proxyURI, net.DefaultResolver, timeout, timeout)
	assert.Nil(t, ret)
	require.ErrorIs(t, err, syscall.ECONNREFUSED)
}

func TestNewSMTPClientFailed(t *testing.T) {
	domain := "zzzz171777.com"
	timeout := 5 * time.Second
	ret, _, err := newSMTPClient(domain, "", net.DefaultResolver, timeout, timeout)
	assert.Nil(t, ret)
	assert.Error(t, err)
}

func TestDialSMTPFailed_NoPortIsConfigured(t *testing.T) {
	disposableDomain := "zzzz1717.com"
	timeout := 5 * time.Second
	ret, err := dialSMTP(disposableDomain, "", net.DefaultResolver, timeout, timeout)
	assert.Nil(t, ret)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing port")
}

func TestDialSMTPFailed_NoSuchHost(t *testing.T) {
	disposableDomain := "zzzzyyyyaaa123.com:25"
	timeout := 5 * time.Second
	ret, err := dialSMTP(disposableDomain, "", net.DefaultResolver, timeout, timeout)
	assert.Nil(t, ret)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no such host")
}

func TestDialSMTP_WithCustomResolver(t *testing.T) {
	domain := "github.com:25"
	timeout := 5 * time.Second
	wantErr := errors.New("custom resolver dial invoked")

	var called atomic.Bool
	customResolver := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			called.Store(true)
			return nil, wantErr
		},
	}

	ret, err := dialSMTP(domain, "", customResolver, timeout, timeout)
	assert.Nil(t, ret)
	assert.True(t, called.Load())
	require.ErrorContains(t, err, wantErr.Error())
}

// The counterpart to the test above: SOCKS5 hands the target hostname to the
// proxy to resolve remotely, so the custom resolver must not be consulted on
// the proxy path. This pins the behaviour the README documents; if the proxy
// path ever gains resolver support, this test should fail and prompt a docs
// update. The proxy address is an IP so that resolving the proxy's own name
// cannot influence the assertion.
func TestDialSMTP_ProxyBypassesCustomResolver(t *testing.T) {
	domain := "github.com:25"
	proxyURI := "socks5://127.0.0.1:1080"
	timeout := 5 * time.Second

	var called atomic.Bool
	customResolver := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			called.Store(true)
			return nil, errors.New("custom resolver dial invoked")
		},
	}

	ret, err := dialSMTP(domain, proxyURI, customResolver, timeout, timeout)
	assert.Nil(t, ret)
	require.Error(t, err)
	assert.False(t, called.Load())
}

// Pins the schemes x/net/proxy actually resolves; socks4 and socks4a never worked.
func TestEstablishProxyConnection_Schemes(t *testing.T) {
	// Nothing listens here: a resolved scheme reaches a refused connection.
	const proxyAddr = "127.0.0.1:1"

	for _, scheme := range []string{"socks5", "socks5h"} {
		t.Run(scheme+" is supported", func(tt *testing.T) {
			_, err := establishProxyConnection("example.com:25", scheme+"://"+proxyAddr, time.Second)
			require.Error(tt, err)
			require.ErrorContains(tt, err, "socks connect")
		})
	}

	for _, scheme := range []string{"socks4", "socks4a", "http", "https"} {
		t.Run(scheme+" is not supported", func(tt *testing.T) {
			_, err := establishProxyConnection("example.com:25", scheme+"://"+proxyAddr, time.Second)
			require.ErrorContains(tt, err, "proxy: unknown scheme: "+scheme)
		})
	}
}

type dialerWithoutContext struct{ err error }

func (d dialerWithoutContext) Dial(_, _ string) (net.Conn, error) { return nil, d.err }

func TestEstablishProxyConnection_DialerWithoutContext(t *testing.T) {
	wantErr := errors.New("registered dialer reached")
	proxy.RegisterDialerType("emailverifiertest", func(*url.URL, proxy.Dialer) (proxy.Dialer, error) {
		return dialerWithoutContext{err: wantErr}, nil
	})

	// Used to panic on the proxy.ContextDialer assertion.
	conn, err := establishProxyConnection("example.com:25", "emailverifiertest://127.0.0.1:1080", time.Second)
	assert.Nil(t, conn)
	require.ErrorIs(t, err, wantErr)
}

type slowDialer struct{ delay time.Duration }

func (d slowDialer) Dial(_, _ string) (net.Conn, error) {
	time.Sleep(d.delay)
	return nil, errors.New("slow dialer finished")
}

func TestEstablishProxyConnection_DialerWithoutContextTimesOut(t *testing.T) {
	proxy.RegisterDialerType("emailverifierslow", func(*url.URL, proxy.Dialer) (proxy.Dialer, error) {
		return slowDialer{delay: 2 * time.Second}, nil
	})

	start := time.Now()
	conn, err := establishProxyConnection("example.com:25", "emailverifierslow://127.0.0.1:1080", 50*time.Millisecond)
	assert.Nil(t, conn)
	require.ErrorIs(t, err, context.DeadlineExceeded)
	assert.Less(t, time.Since(start), time.Second)
}

func TestPreferredDialError(t *testing.T) {
	network := errors.New("dial tcp 1.2.3.4:25: connect: connection refused")
	dns := &net.DNSError{Err: "no such host", Name: "mx.example.invalid", IsNotFound: true}
	reply := &textproto.Error{Code: 550, Msg: "5.7.1 Service unavailable, Client host blocked using Spamhaus."}
	rendered := errors.New("421 4.7.0 Too many concurrent connections")

	t.Run("prefers a server reply over a network failure", func(tt *testing.T) {
		// the network failure arrives first because it fails fastest
		assert.Same(tt, reply, preferredDialError([]error{network, reply}))
	})

	t.Run("prefers a server reply over a DNS failure", func(tt *testing.T) {
		assert.Same(tt, reply, preferredDialError([]error{dns, reply}))
	})

	t.Run("recognises a reply that was already rendered to a string", func(tt *testing.T) {
		assert.Same(tt, rendered, preferredDialError([]error{network, rendered}))
	})

	t.Run("falls back to the first error when no server replied", func(tt *testing.T) {
		assert.Same(tt, network, preferredDialError([]error{network, dns}))
	})

	t.Run("single error", func(tt *testing.T) {
		assert.Same(tt, network, preferredDialError([]error{network}))
	})
}

func TestHasSMTPReply(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want bool
	}{
		{"textproto error", &textproto.Error{Code: 550, Msg: "no such user"}, true},
		{"rendered reply", errors.New("452 4.5.3 Recipients belong to multiple regions"), true},
		{"connection refused", errors.New("dial tcp 1.2.3.4:25: connect: connection refused"), false},
		{"dns miss", &net.DNSError{Err: "no such host"}, false},
		{"too short to carry a code", errors.New("no"), false},
		{"leading digits that are not a reply code", errors.New("103.26.221.42 unreachable"), false},
		{"no MX records", errors.New("No MX records found"), false},
	}
	for _, c := range cases {
		test := c
		t.Run(test.name, func(tt *testing.T) {
			assert.Equal(tt, test.want, hasSMTPReply(test.err))
		})
	}
}

// The API-verifier branch returns before our own catch-all probe runs, so the
// outcome has to come from what the verifier itself reported. Leaving it at
// catchAllNotRun would turn a verifier's CatchAll: true into reachable "no".
func TestCatchAllFromResult(t *testing.T) {
	assert.Equal(t, catchAllAccepted, catchAllFromResult(&SMTP{CatchAll: true}))
	assert.Equal(t, catchAllRefused, catchAllFromResult(&SMTP{CatchAll: false}))
	assert.Equal(t, catchAllRefused, catchAllFromResult(nil))
}

// GenerateRandomEmail feeds the catch-all probe: every SMTP exchange sends RCPT
// TO for this address and reads the reply as "does this domain accept anything".
// Nothing asserted its output before, so its shape was only exercised
// incidentally through CheckSMTP, which needs a live mail server.
func TestGenerateRandomEmail(t *testing.T) {
	const domain = "example.com"

	got := GenerateRandomEmail(domain)

	local, host, found := strings.Cut(got, "@")
	require.True(t, found, "address must contain an @: %q", got)
	assert.Equal(t, domain, host, "the domain is passed through unchanged")
	assert.Len(t, local, 32, "the local part is a fixed 32 characters")

	// The probe is only meaningful if the address cannot collide with a real
	// mailbox. A short or guessable local part risks matching one, which would
	// read as "this domain is catch-all" and mark every address on it
	// unverifiable.
	assert.Regexp(t, `^[a-z0-9]+$`, local, "local part must stay inside alphanumeric")
}

func TestGenerateRandomEmail_DoesNotRepeat(t *testing.T) {
	// This asserts distinctness, not unpredictability: a fixed sequence of
	// distinct values would pass. It is still worth pinning, because a repeat
	// is what would let the probe collide with a real mailbox -- the server then
	// answers 250 for an address that does not exist, which reads as "this domain
	// accepts anything" and marks every address on it unverifiable.
	seen := make(map[string]struct{}, 100)
	for range 100 {
		seen[GenerateRandomEmail("example.com")] = struct{}{}
	}
	assert.Greater(t, len(seen), 90, "100 draws should almost all be distinct")
}

func TestGenerateRandomEmail_KeepsTheDomainVerbatim(t *testing.T) {
	// Subdomains and mixed case must survive, because the address is handed
	// straight to RCPT TO and a mangled domain would be rejected for the wrong
	// reason.
	for _, domain := range []string{"mail.example.co.uk", "Example.COM", "xn--fsq.example"} {
		got := GenerateRandomEmail(domain)
		assert.True(t, strings.HasSuffix(got, "@"+domain), "%q should end with @%s", got, domain)
	}
}

// fakeDNS is a local UDP DNS server that answers MX questions from a table and
// counts what it was asked. It exists so the MX-then-dial ordering can be tested
// without a network: the four newSMTPClient tests that predate these all pass
// net.DefaultResolver, which means they only assert anything when the machine
// has working DNS and an SMTP server that answers.
//
// The name is deliberately not "helpers": there is no helpers.go in this package,
// and a <something>_test.go file that has no matching source file reads as
// misplaced. These helpers stay next to the only tests that use them.
type fakeDNS struct {
	addr    string
	queries atomic.Int32
	mx      []string // hosts to answer MX questions with; empty means answer with none
	rcode   uint8
}

func newFakeDNS(t *testing.T, mx []string, rcode dnsmessage.RCode) *fakeDNS {
	t.Helper()

	var lc net.ListenConfig
	pc, err := lc.ListenPacket(context.Background(), "udp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { _ = pc.Close() })

	f := &fakeDNS{addr: pc.LocalAddr().String(), mx: mx, rcode: uint8(rcode)}

	go func() {
		buf := make([]byte, 512)
		for {
			n, from, err := pc.ReadFrom(buf)
			if err != nil {
				return
			}
			var parser dnsmessage.Parser
			header, err := parser.Start(buf[:n])
			if err != nil {
				continue
			}
			questions, err := parser.AllQuestions()
			if err != nil {
				continue
			}

			builder := dnsmessage.NewBuilder(nil, dnsmessage.Header{
				ID: header.ID, Response: true, Authoritative: true,
				RCode: dnsmessage.RCode(f.rcode),
			})
			builder.EnableCompression()
			if err := builder.StartQuestions(); err != nil {
				continue
			}
			for _, q := range questions {
				_ = builder.Question(q)
			}
			if err := builder.StartAnswers(); err != nil {
				continue
			}
			for _, q := range questions {
				// The MX hostnames also need A records, or the dial cannot
				// resolve them and fails in DNS rather than at the socket --
				// which would make the port assertion below untestable.
				if q.Type == dnsmessage.TypeA && f.hostsA(q.Name.String()) {
					_ = builder.AResource(
						dnsmessage.ResourceHeader{
							Name:  q.Name,
							Type:  dnsmessage.TypeA,
							Class: dnsmessage.ClassINET,
							TTL:   60,
						},
						dnsmessage.AResource{A: [4]byte{127, 0, 0, 1}},
					)
					continue
				}
				if q.Type != dnsmessage.TypeMX {
					continue
				}
				f.queries.Add(1)
				for i, host := range f.mx {
					name, err := dnsmessage.NewName(host)
					if err != nil {
						continue
					}
					_ = builder.MXResource(
						dnsmessage.ResourceHeader{
							Name:  q.Name,
							Type:  dnsmessage.TypeMX,
							Class: dnsmessage.ClassINET,
							TTL:   60,
						},
						dnsmessage.MXResource{Pref: uint16(i), MX: name},
					)
				}
			}
			response, err := builder.Finish()
			if err != nil {
				continue
			}
			_, _ = pc.WriteTo(response, from)
		}
	}()

	return f
}

// hostsA reports whether name is one of the MX hosts this server answers for.
func (f *fakeDNS) hostsA(name string) bool {
	for _, host := range f.mx {
		if host == name {
			return true
		}
	}
	return false
}

// resolver points a *net.Resolver at the fake server. PreferGo keeps it on the
// pure-Go resolver, which is the one that respects Dial; with cgo the lookup can
// go to libc and ignore this entirely.
func (f *fakeDNS) resolver() *net.Resolver {
	return &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, _ string) (net.Conn, error) {
			var d net.Dialer
			return d.DialContext(ctx, network, f.addr)
		},
	}
}

// The MX lookup is what decides the address that gets dialled, so it has to
// happen first. Pinning the order matters because moving the dial ahead of the
// lookup would still produce an error -- just a less useful one -- and that
// regression would be invisible to every test that predates these.
func TestNewSMTPClientLooksUpMXBeforeDialing(t *testing.T) {
	f := newFakeDNS(t, []string{"mail.example.com."}, dnsmessage.RCodeSuccess)
	timeout := 2 * time.Second

	client, selected, err := newSMTPClient("example.com", "", f.resolver(), timeout, timeout)

	assert.Nil(t, client)
	assert.Nil(t, selected)
	// A resolved MX is dialled, and refusing is proof the dial was reached.
	require.Error(t, err)
	require.ErrorIs(t, err, syscall.ECONNREFUSED)
	// At least one MX question: the lookup ran, and the ECONNREFUSED above proves
	// the dial came after it. The count is not pinned to one because the resolver
	// may retry internally, which is its business rather than this function's.
	assert.GreaterOrEqual(t, int(f.queries.Load()), 1, "expected the MX question before the dial")
}

// When the lookup fails, the caller must be told the name could not be resolved.
// Reporting a connection failure here would send people looking at firewalls and
// port 25 for a problem that is in DNS -- the two diagnoses have nothing in
// common, so the error has to stay on the DNS side of the line.
func TestNewSMTPClientMXFailureIsADNSErrorNotADialError(t *testing.T) {
	f := newFakeDNS(t, nil, dnsmessage.RCodeNameError)
	timeout := 2 * time.Second

	client, selected, err := newSMTPClient("example.com", "", f.resolver(), timeout, timeout)

	assert.Nil(t, client)
	assert.Nil(t, selected)
	require.Error(t, err)

	var dnsErr *net.DNSError
	require.ErrorAs(t, err, &dnsErr)
	assert.True(t, dnsErr.IsNotFound, "expected a not-found DNS error, got %#v", dnsErr)
	require.NotErrorIs(t, err, syscall.ECONNREFUSED)
	// Same reasoning as above, and here it matters more: a resolver that retries a
	// failed lookup will legitimately ask more than once.
	assert.GreaterOrEqual(t, int(f.queries.Load()), 1, "expected at least one MX question")
}

// The address that gets dialled is the MX host plus smtpPort -- the constant, not
// anything the caller supplied. This is the behaviour behind the note on
// newSMTPClient's signature: the port is not a parameter, so injecting one at the
// call site cannot change it. Asserting the assembled address keeps that true
// even after the port becomes configurable, as long as the default stays 25.
func TestNewSMTPClientDialsTheResolvedMXHostOnTheSMTPPort(t *testing.T) {
	f := newFakeDNS(t, []string{"mail.example.com."}, dnsmessage.RCodeSuccess)
	timeout := 2 * time.Second

	_, _, err := newSMTPClient("example.com", "", f.resolver(), timeout, timeout)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "dial tcp 127.0.0.1"+smtpPort,
		"expected the dial to target the resolved MX host on smtpPort")
}
