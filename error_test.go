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
func TestParseError_Code400_ReportedVerbatim(t *testing.T) {
	errStr := "400"
	err := errors.New(errStr)
	le := ParseSMTPError(err)

	require.NotNil(t, le)
	assert.Equal(t, errStr, le.Message)
	assert.Equal(t, errStr, le.Details)
	assert.NotPanics(t, func() { _ = le.Error() })
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
	require.True(t, errors.As(error(le), &dnsErr))
	assert.True(t, dnsErr.IsNotFound)
	assert.True(t, errors.Is(error(le), cause))
}

func TestLookupError_UnwrapNilSafe(t *testing.T) {
	var le *LookupError
	assert.Nil(t, le.Unwrap())
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
