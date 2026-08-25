package emailverifier

import (
	"context"
	"errors"
	"net"
	"net/textproto"
	"strings"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	assert.NoError(t, err)
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
	assert.NoError(t, err)
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
	assert.NoError(t, err)
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
	assert.NoError(t, err)
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
	assert.NoError(t, err)
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
	assert.NoError(t, err)
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
	assert.NoError(t, err)
	assert.Equal(t, &expected, smtp)
}

func TestCheckSMTP_DisabledSMTPCheck(t *testing.T) {
	domain := "github.com"

	verifier.DisableSMTPCheck()
	smtp, err := verifier.CheckSMTP(domain, "username")
	verifier.EnableSMTPCheck()

	assert.NoError(t, err)
	assert.Nil(t, smtp)
}

func TestCheckSMTPOK_HostNotExists(t *testing.T) {
	domain := "notExistHost.com"

	smtp, err := verifier.CheckSMTP(domain, "")
	assert.Error(t, err, ErrNoSuchHost)
	assert.Equal(t, &SMTP{}, smtp)
}

func TestNewSMTPClientOK(t *testing.T) {
	domain := "gmail.com"
	timeout := 5 * time.Second
	ret, _, err := newSMTPClient(domain, "", net.DefaultResolver, timeout, timeout)
	assert.NotNil(t, ret)
	assert.Nil(t, err)
}

func TestNewSMTPClientFailed_WithInvalidProxy(t *testing.T) {
	domain := "gmail.com"
	proxyURI := "socks5://user:password@127.0.0.1:1080?timeout=5s"
	timeout := 5 * time.Second
	ret, _, err := newSMTPClient(domain, proxyURI, net.DefaultResolver, timeout, timeout)
	assert.Nil(t, ret)
	assert.Error(t, err, syscall.ECONNREFUSED)
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
	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "missing port"))
}

func TestDialSMTPFailed_NoSuchHost(t *testing.T) {
	disposableDomain := "zzzzyyyyaaa123.com:25"
	timeout := 5 * time.Second
	ret, err := dialSMTP(disposableDomain, "", net.DefaultResolver, timeout, timeout)
	assert.Nil(t, ret)
	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "no such host"))
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
