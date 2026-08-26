package emailverifier

import (
	"errors"
	"net"
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
