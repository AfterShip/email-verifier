package emailverifier

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
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
	assert.Equal(t, &LookupError{Details: errStr, Message: errStr}, le)
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
	assert.Equal(t, &LookupError{Details: errStr, Message: errStr}, le)
}

func TestParseError_Code550(t *testing.T) {
	errStr := "550"
	err := errors.New(errStr)
	le := ParseSMTPError(err)

	assert.Equal(t, ErrServerUnavailable, le.Message)
	assert.Equal(t, err.Error(), le.Details)
}

func TestParseError_Code400_Nil(t *testing.T) {
	errStr := "400"
	err := errors.New(errStr)
	le := ParseSMTPError(err)

	assert.Equal(t, (*LookupError)(nil), le)
}

func TestParseError_Code401(t *testing.T) {
	errStr := "401"
	err := errors.New(errStr)
	le := ParseSMTPError(err)

	assert.Equal(t, &LookupError{Details: errStr, Message: errStr}, le)
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
