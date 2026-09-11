package emailverifier

import (
	"errors"
	"net"
	"net/textproto"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParse550RCPTError(t *testing.T) {
	err := errors.New("550 This mailbox does not exist")
	le := ParseSMTPError(err)
	assert.Equal(t, ErrServerUnavailable, le.Message)
	assert.Equal(t, err.Error(), le.Details)
}

func TestParse550BlockedRCPTError(t *testing.T) {
	err := errors.New("550 spamhaus")
	le := ParseSMTPError(err)
	assert.Equal(t, ErrBlocked, le.Message)
	assert.Equal(t, err.Error(), le.Details)
}

func TestParseConnectMailExchangerError(t *testing.T) {
	err := errors.New("Timeout connecting to mail-exchanger")
	le := ParseSMTPError(err)
	assert.Equal(t, ErrTimeout, le.Message)
	assert.Equal(t, err.Error(), le.Details)
}

func TestParseNoMxRecordsFoundError(t *testing.T) {
	errStr := "No MX records found"
	err := errors.New(errStr)
	le := ParseSMTPError(err)
	assert.Equal(t, errStr, le.Message)
	assert.Equal(t, errStr, le.Details)
}

func TestParseFullInBoxError(t *testing.T) {
	errStr := "452 full Inbox"
	err := errors.New(errStr)
	le := ParseSMTPError(err)

	assert.Equal(t, ErrFullInbox, le.Message)
	assert.Equal(t, err.Error(), le.Details)
}

func TestParseDailSMTPServerError(t *testing.T) {
	errStr := "Unexpected response dialing SMTP server"
	err := errors.New(errStr)
	le := ParseSMTPError(err)
	assert.Equal(t, errStr, le.Message)
	assert.Equal(t, errStr, le.Details)
}

func TestParseError_Code550(t *testing.T) {
	errStr := "550"
	err := errors.New(errStr)
	le := ParseSMTPError(err)

	assert.Equal(t, ErrServerUnavailable, le.Message)
	assert.Equal(t, err.Error(), le.Details)
}

// A status line that does not itself indicate a failure is still reported,
// because ParseSMTPError is only ever reached with a non-nil error. Returning
// nil used to hand callers a non-nil error interface wrapping a nil
// *LookupError, and calling Error on that panics.
func TestParseError_UnclassifiedIsReportedVerbatim(t *testing.T) {
	// Anything whose status line parses to 400 or below takes this branch. Real
	// servers do not send a success code as an error, but the guarantee is about
	// the function's contract: it is only ever called with a non-nil error, so it
	// must not discard one.
	for _, reply := range []string{"200 OK", "300 Redirect", "399", "400"} {
		test := reply
		t.Run(test, func(tt *testing.T) {
			cause := errors.New(test)

			le := ParseSMTPError(cause)

			require.NotNil(tt, le)
			assert.Equal(tt, test, le.Message)
			assert.Equal(tt, test, le.Details)
			assert.NotPanics(tt, func() { _ = le.Error() })
			// the cause has to survive on this branch too, not just the
			// classified one
			assert.ErrorIs(tt, error(le), cause)
		})
	}
}

func TestParseSMTPError_NilInput(t *testing.T) {
	assert.Nil(t, ParseSMTPError(nil))
}

func TestParseSMTPError_UnwrapsToCause(t *testing.T) {
	cause := &net.DNSError{Err: "no such host", Name: "example.invalid", IsNotFound: true}

	le := ParseSMTPError(cause)
	require.NotNil(t, le)
	assert.Equal(t, ErrNoSuchHost, le.Message)

	// the point of wrapping: callers can inspect the underlying error by type
	// instead of matching on Details
	var dnsErr *net.DNSError
	require.ErrorAs(t, error(le), &dnsErr)
	assert.True(t, dnsErr.IsNotFound)
	assert.ErrorIs(t, error(le), cause)
}

func TestLookupError_UnwrapNilSafe(t *testing.T) {
	var le *LookupError
	assert.NoError(t, le.Unwrap())
}

func TestParseError_Code401(t *testing.T) {
	errStr := "401"
	err := errors.New(errStr)
	le := ParseSMTPError(err)

	assert.Equal(t, errStr, le.Message)
	assert.Equal(t, errStr, le.Details)
}

func TestParseError_Code421(t *testing.T) {
	errStr := "421"
	err := errors.New(errStr)
	le := ParseSMTPError(err)

	assert.Equal(t, ErrTryAgainLater, le.Message)
	assert.Equal(t, err.Error(), le.Details)
}

func TestParseError_Code450(t *testing.T) {
	errStr := "450"
	err := errors.New(errStr)
	le := ParseSMTPError(err)

	assert.Equal(t, ErrMailboxBusy, le.Message)
	assert.Equal(t, err.Error(), le.Details)
}

func TestParseError_Code451(t *testing.T) {
	errStr := "451"
	err := errors.New(errStr)
	le := ParseSMTPError(err)

	assert.Equal(t, ErrExceededMessagingLimits, le.Message)
	assert.Equal(t, err.Error(), le.Details)
}

func TestParseError_Code452(t *testing.T) {
	errStr := "452"
	err := errors.New(errStr)
	le := ParseSMTPError(err)

	assert.Equal(t, ErrTooManyRCPT, le.Message)
	assert.Equal(t, err.Error(), le.Details)
}

func TestParseError_Code503(t *testing.T) {
	errStr := "503"
	err := errors.New(errStr)
	le := ParseSMTPError(err)

	assert.Equal(t, ErrNeedMAILBeforeRCPT, le.Message)
	assert.Equal(t, err.Error(), le.Details)
}

func TestParseError_Code551(t *testing.T) {
	errStr := "551"
	err := errors.New(errStr)
	le := ParseSMTPError(err)

	assert.Equal(t, ErrRCPTHasMoved, le.Message)
	assert.Equal(t, err.Error(), le.Details)
}

func TestParseError_Code552(t *testing.T) {
	errStr := "552"
	err := errors.New(errStr)
	le := ParseSMTPError(err)

	assert.Equal(t, ErrFullInbox, le.Message)
	assert.Equal(t, err.Error(), le.Details)
}

func TestParseError_Code553(t *testing.T) {
	errStr := "553"
	err := errors.New(errStr)
	le := ParseSMTPError(err)

	assert.Equal(t, ErrNoRelay, le.Message)
	assert.Equal(t, err.Error(), le.Details)
}

func TestParseError_Code554(t *testing.T) {
	errStr := "554"
	err := errors.New(errStr)
	le := ParseSMTPError(err)

	assert.Equal(t, ErrNotAllowed, le.Message)
	assert.Equal(t, err.Error(), le.Details)
}

func TestParseError_basicErr_timeout(t *testing.T) {
	errStr := "559 timeout"
	err := errors.New(errStr)
	le := ParseSMTPError(err)

	assert.Equal(t, ErrTimeout, le.Message)
	assert.Equal(t, err.Error(), le.Details)
}

func TestParseError_basicErr_blocked(t *testing.T) {
	errStr := "559 blocked"
	err := errors.New(errStr)
	le := ParseSMTPError(err)

	assert.Equal(t, ErrBlocked, le.Message)
	assert.Equal(t, err.Error(), le.Details)
}

// Replies captured 2026-09-10 by driving the exchange directly, so the table
// doubles as a record of what enhanced status codes look like in the wild.
// Verbatim, session identifiers included; only our egress IP is replaced.
//
// Each is a *textproto.Error because that is what net/smtp returns. A table of
// plain strings passed while the feature was broken for every real reply.
func TestParseSMTPError_EnhancedCode(t *testing.T) {
	cases := []struct {
		name     string
		err      error
		enhanced string
	}{
		{
			// Multi-line: textproto joins the lines with \n, each repeating the code.
			name: "gmail, recipient unknown, RCPT",
			err: &textproto.Error{Code: 550, Msg: "5.1.1 The email account that you tried to reach does not exist. Please try\n" +
				"5.1.1 double-checking the recipient's email address for typos or\n" +
				"5.1.1 unnecessary spaces. For more information, go to\n" +
				"5.1.1  https://support.google.com/mail/?p=NoSuchUser 41be03b00d2f7-cc4c687ec03si136868a12.335 - gsmtp"},
			enhanced: "5.1.1",
		},
		{
			// X.5.x for an unknown mailbox, not X.1.x: the subject is not a verdict.
			name:     "microsoft, recipient unknown, RCPT",
			err:      &textproto.Error{Code: 550, Msg: "5.5.0 Requested action not taken: mailbox unavailable (S2017062302). [SJ1PEPF000037A2.namprd04.prod.outlook.com 2026-09-10T17:20:46.631Z 08DF0B1332F1DE15]"},
			enhanced: "5.5.0",
		},
		{
			// Rejected at MAIL FROM: no recipient had been named.
			name:     "yahoo, sender fails FCrDNS, MAIL FROM",
			err:      &textproto.Error{Code: 550, Msg: "5.7.25 Forward-confirmed reverse DNS failed tnmpmscs"},
			enhanced: "5.7.25",
		},
		{
			name:     "zoho, recipient unknown, RCPT",
			err:      &textproto.Error{Code: 550, Msg: "5.1.1 User does not exist - <rcpwt0gy04lubc4qmkpppxr128xgnaef@zoho.com>"},
			enhanced: "5.1.1",
		},
		{
			// No enhanced code despite a 550 about the recipient. Egress IP replaced.
			name:     "qq sends no enhanced code, RCPT",
			err:      &textproto.Error{Code: 550, Msg: "Mailbox not found. http://service.mail.qq.com/detail/122/169 [MAW77ic63OtZB8WdcrvY/gIHPvUvsXupE8bE5roaPk+ab2Lvf59Ok0394SdGexrgVA== IP: 198.51.100.1]"},
			enhanced: "",
		},
		{
			name:     "netease sends no enhanced code, RCPT",
			err:      &textproto.Error{Code: 550, Msg: "User not found: lyta8y5m8nenxwdkg2vybjw2n2cjg602@163.com"},
			enhanced: "",
		},
		{
			// Earlier run. The only transient reply captured, and for a real mailbox.
			name:     "microsoft, transient, RCPT",
			err:      &textproto.Error{Code: 452, Msg: "4.5.3 Recipients belong to multiple regions ATTR38 [AMS1EPF0000008F.eurprd05.prod.outlook.com]"},
			enhanced: "4.5.3",
		},
		{
			// Carries a bare IPv4 address the pattern must not mistake for a code.
			name:     "sender blocklisted, reply contains an IP address",
			err:      &textproto.Error{Code: 550, Msg: "5.7.1 Service unavailable, Client host [198.51.100.1] blocked using Spamhaus."},
			enhanced: "5.7.1",
		},
		{
			// X.7.x -- reserved for policy -- for an unknown recipient. With the
			// case below, one code meaning two opposite things.
			name:     "yandex, recipient unknown, answered with 5.7.1",
			err:      &textproto.Error{Code: 550, Msg: "5.7.1 No such user! 1789063127-kwTVGUHdeiE0-huj9x1Wr"},
			enhanced: "5.7.1",
		},
		{
			// The same code, meaning our IP is blocklisted. No recipient was judged.
			name:     "apple, sender blocklisted, answered with 5.7.1",
			err:      &textproto.Error{Code: 550, Msg: "5.7.1 Mail from IP 198.51.100.1 was rejected due to listing in Spamhaus SBL. For details please see http://www.spamhaus.org/query/bl?ip=198.51.100.1"},
			enhanced: "5.7.1",
		},
		{
			// The default HelloName rejected at EHLO. 501 is not a code
			// parseSMTPError switches on; see the Message test below.
			name:     "fastmail rejects EHLO localhost",
			err:      &textproto.Error{Code: 501, Msg: "5.7.1 <localhost>: Helo command rejected: HELO string 'localhost' not accepted, Use a real hostname/ip"},
			enhanced: "5.7.1",
		},
		{
			// Same rejection, but Hello returned nil and it surfaced at RCPT.
			name:     "tuta rejects EHLO localhost, surfacing at RCPT",
			err:      &textproto.Error{Code: 504, Msg: "5.5.2 <localhost>: Helo command rejected: need fully-qualified hostname"},
			enhanced: "5.5.2",
		},
		{
			// Refused at the greeting: multi-line, no enhanced code.
			name: "united internet refuses the connection",
			err: &textproto.Error{Code: 554, Msg: "gmx.net (mxgmx009) Nemesis ESMTP Service not available\n" +
				"No SMTP service\n" +
				"IP address is block listed.\n" +
				"For explanation visit https://postmaster.gmx.net/en/case?c=r0303&i=ip&v=198.51.100.1&r=1N6a0i-1wlEnE2roU-012Max"},
			enhanced: "",
		},
		{
			// The default FromEmail: example.org publishes v=spf1 -all.
			name:     "aliyun rejects the default sender on spf",
			err:      &textproto.Error{Code: 554, Msg: "Reject by behaviour spam at Rcpt State(Connection IP address:198.51.100.1)ANTISPAM_BAT[01201311R846a, maildocker-behaviorspam033068209079]: spf check failedCONTINUE"},
			enhanced: "",
		},
		{
			// A rate limit wearing a 550, with no enhanced code to say so.
			name:     "sina rate-limits with a 550",
			err:      &textproto.Error{Code: 550, Msg: "Too many recipients."},
			enhanced: "",
		},
		{
			name:     "dns failure carries no reply",
			err:      &net.DNSError{Err: "no such host", Name: "icloud.com", IsNotFound: true},
			enhanced: "",
		},
		{
			// Code and Msg are separate fields: a reply code we cannot parse says
			// nothing about Msg. ParseSMTPError is exported and takes any error.
			name:     "unparseable reply code, enhanced code still read",
			err:      &textproto.Error{Code: 1234, Msg: "5.1.1 no such user"},
			enhanced: "5.1.1",
		},
		{
			name:     "3xx reply, no enhanced code of a class we recognise",
			err:      &textproto.Error{Code: 354, Msg: "3.0.0 start mail input"},
			enhanced: "",
		},
		{
			name:     "connection failure, rendered as a string",
			err:      errors.New("dial tcp 1.2.3.4:25: connect: connection refused"),
			enhanced: "",
		},
		{
			name:     "reply rendered as a string",
			err:      errors.New("550 5.1.1 does not exist"),
			enhanced: "5.1.1",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(tt *testing.T) {
			le := ParseSMTPError(c.err)
			require.NotNil(tt, le)
			assert.Equal(tt, c.enhanced, le.EnhancedCode())
		})
	}
}

func TestLookupError_EnhancedCodeNilSafe(t *testing.T) {
	var le *LookupError
	assert.Empty(t, le.EnhancedCode())
}

// Pins current behaviour rather than endorsing it: Details, and Message for a
// reply code parseSMTPError does not switch on, are built from the %q rendering,
// so callers see escaped quotes and a literal \n. Details is serialised, so this
// reaches everyone. Changing it wants its own review; this stops it drifting.
func TestParseSMTPError_RenderedReplyLeaksIntoMessageAndDetails(t *testing.T) {
	t.Run("Details always carries the rendered reply", func(tt *testing.T) {
		le := ParseSMTPError(&textproto.Error{Code: 550, Msg: "Too many recipients."})
		assert.Equal(tt, ErrServerUnavailable, le.Message)
		assert.Equal(tt, `550 "Too many recipients."`, le.Details)
	})

	t.Run("an unswitched reply code leaves it in Message too", func(tt *testing.T) {
		le := ParseSMTPError(&textproto.Error{Code: 501, Msg: "5.7.1 <localhost>: Helo command rejected"})
		assert.Equal(tt, `501 "5.7.1 <localhost>: Helo command rejected"`, le.Message)
		assert.Equal(tt, le.Message, le.Details)
	})

	t.Run("EnhancedCode is unaffected, being read from Msg", func(tt *testing.T) {
		le := ParseSMTPError(&textproto.Error{Code: 501, Msg: "5.7.1 <localhost>: Helo command rejected"})
		assert.Equal(tt, "5.7.1", le.EnhancedCode())
	})
}
