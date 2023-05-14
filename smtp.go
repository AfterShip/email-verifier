package emailverifier

import (
	"errors"
	"fmt"
	"math/rand"
	"net"
	"net/smtp"
	"net/url"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/proxy"
)

// SMTP stores all information for SMTP verification lookup
type SMTP struct {
	HostExists  bool `json:"host_exists"` // is the host exists?
	FullInbox   bool `json:"full_inbox"`  // is the email account's inbox full?
	CatchAll    bool `json:"catch_all"`   // does the domain have a catch-all email address?
	Deliverable bool `json:"deliverable"` // can send an email to the email server?
	Disabled    bool `json:"disabled"`    // is the email blocked or disabled by the provider?
}

// CheckSMTP performs an email verification on the passed domain via SMTP
//   - the domain is the passed email domain
//   - username is used to check the deliverability of specific email address,
//
// if server is catch-all server, username will not be checked
func (v *Verifier) CheckSMTP(domain, username string) (*SMTP, error) {
	if !v.smtpCheckEnabled {
		return nil, nil
	}

	var ret SMTP
	var err error
	email := fmt.Sprintf("%s@%s", username, domain)

	// Dial any SMTP server that will accept a connection
	client, mx, err := newSMTPClient(domain, v.proxyURI)
	if err != nil {
		return &ret, ParseSMTPError(err)
	}

	// Defer quit the SMTP connection
	defer client.Close()

	// Check by api when enabled and host recognized.
	for _, apiVerifier := range v.apiVerifiers {
		if apiVerifier.isSupported(strings.ToLower(mx.Host)) {
			return apiVerifier.check(domain, username)
		}
	}

	// Sets the HELO/EHLO hostname
	if err = client.Hello(v.helloName); err != nil {
		return &ret, ParseSMTPError(err)
	}

	// Sets the from email
	if err = client.Mail(v.fromEmail); err != nil {
		return &ret, ParseSMTPError(err)
	}

	// Host exists if we've successfully formed a connection
	ret.HostExists = true

	// Default sets catch-all to true
	ret.CatchAll = true

	if v.catchAllCheckEnabled {
		// Checks the deliver ability of a randomly generated address in
		// order to verify the existence of a catch-all and etc.
		randomEmail := GenerateRandomEmail(domain)
		if err = client.Rcpt(randomEmail); err != nil {
			if e := ParseSMTPError(err); e != nil {
				switch e.Message {
				case ErrFullInbox:
					ret.FullInbox = true
				case ErrNotAllowed:
					ret.Disabled = true
				// If The client typically receives a `550 5.1.1` code as a reply to RCPT TO command,
				// In most cases, this is because the recipient address does not exist.
				case ErrServerUnavailable:
					ret.CatchAll = false
				default:

				}

			}
		}

		// If the email server is a catch-all email server,
		// no need to calibrate deliverable on a specific user
		if ret.CatchAll {
			return &ret, nil
		}
	}

	// If no username provided,
	// no need to calibrate deliverable on a specific user
	if username == "" {
		return &ret, nil
	}

	if err = client.Rcpt(email); err == nil {
		ret.Deliverable = true
	}

	return &ret, nil
}

// newSMTPClient generates a new available SMTP client
func newSMTPClient(domain, proxyURI string) (*smtp.Client, *net.MX, error) {
	domain = domainToASCII(domain)
	mxRecords, err := net.LookupMX(domain)
	if err != nil {
		return nil, nil, err
	}

	if len(mxRecords) == 0 {
		return nil, nil, errors.New("No MX records found")
	}
	// Create a channel for receiving response from
	ch := make(chan interface{}, 1)
	selectedMXCh := make(chan *net.MX, 1)

	// Done indicates if we're still waiting on dial responses
	var done bool

	// mutex for data race
	var mutex sync.Mutex

	// Attempt to connect to all SMTP servers concurrently
	for i, r := range mxRecords {
		addr := r.Host + smtpPort
		index := i
		go func() {
			c, err := dialSMTP(addr, proxyURI)
			if err != nil {
				if !done {
					ch <- err
				}
				return
			}

			// Place the client on the channel or close it
			mutex.Lock()
			switch {
			case !done:
				done = true
				ch <- c
				selectedMXCh <- mxRecords[index]
			default:
				c.Close()
			}
			mutex.Unlock()
		}()
	}

	// Collect errors or return a client
	var errs []error
	for {
		res := <-ch
		switch r := res.(type) {
		case *smtp.Client:
			return r, <-selectedMXCh, nil
		case error:
			errs = append(errs, r)
			if len(errs) == len(mxRecords) {
				return nil, nil, errs[0]
			}
		default:
			return nil, nil, errors.New("Unexpected response dialing SMTP server")
		}
	}

}

// dialSMTP is a timeout wrapper for smtp.Dial. It attempts to dial an
// SMTP server (socks5 proxy supported) and fails with a timeout if timeout is reached while
// attempting to establish a new connection
func dialSMTP(addr, proxyURI string) (*smtp.Client, error) {
	// Channel holding the new smtp.Client or error
	ch := make(chan interface{}, 1)

	// Dial the new smtp connection
	go func() {
		var conn net.Conn
		var err error

		if proxyURI != "" {
			conn, err = establishProxyConnection(addr, proxyURI)
		} else {
			conn, err = establishConnection(addr)
		}
		if err != nil {
			ch <- err
			return
		}

		host, _, _ := net.SplitHostPort(addr)
		client, err := smtp.NewClient(conn, host)
		if err != nil {
			ch <- err
			return
		}
		ch <- client
	}()

	// Retrieve the smtp client from our client channel or timeout
	select {
	case res := <-ch:
		switch r := res.(type) {
		case *smtp.Client:
			return r, nil
		case error:
			return nil, r
		default:
			return nil, errors.New("Unexpected response dialing SMTP server")
		}
	case <-time.After(smtpTimeout):
		return nil, errors.New("Timeout connecting to mail-exchanger")
	}
}

// GenerateRandomEmail generates a random email address using the domain passed. Used
// primarily for checking the existence of a catch-all address
func GenerateRandomEmail(domain string) string {
	r := make([]byte, 32)
	for i := 0; i < 32; i++ {
		r[i] = alphanumeric[rand.Intn(len(alphanumeric))]
	}
	return fmt.Sprintf("%s@%s", string(r), domain)

}

// establishConnection connects to the address on the named network address.
func establishConnection(addr string) (net.Conn, error) {
	return net.Dial("tcp", addr)
}

// establishProxyConnection connects to the address on the named network address
// via proxy protocol
func establishProxyConnection(addr, proxyURI string) (net.Conn, error) {
	u, err := url.Parse(proxyURI)
	if err != nil {
		return nil, err
	}
	dialer, err := proxy.FromURL(u, nil)
	if err != nil {
		return nil, err
	}
	return dialer.Dial("tcp", addr)
}
