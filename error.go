package emailverifier

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

const (
	// Standard Errors
	ErrTimeout           = "The connection to the mail server has timed out"
	ErrNoSuchHost        = "Mail server does not exist"
	ErrServerUnavailable = "Mail server is unavailable"
	ErrBlocked           = "Blocked by mail server"

	// RCPT Errors
	ErrTryAgainLater           = "Try again later"
	ErrFullInbox               = "Recipient out of disk space"
	ErrTooManyRCPT             = "Too many recipients"
	ErrNoRelay                 = "Not an open relay"
	ErrMailboxBusy             = "Mailbox busy"
	ErrExceededMessagingLimits = "Messaging limits have been exceeded"
	ErrNotAllowed              = "Not Allowed"
	ErrNeedMAILBeforeRCPT      = "Need MAIL before RCPT"
	ErrRCPTHasMoved            = "Recipient has moved"
)

// LookupError is an MX dns records lookup error
type LookupError struct {
	Message string `json:"message" xml:"message"`
	Details string `json:"details" xml:"details"`

	// cause is the error this was derived from. It is deliberately unexported
	// and not serialised; it exists so that callers can use errors.Is and
	// errors.As to reach the underlying *net.DNSError, *net.OpError or
	// *textproto.Error instead of having to match on Details.
	cause error

	// enhanced is the RFC 3463 enhanced status code the server sent, if any.
	enhanced string
}

// enhancedCodePattern matches an RFC 3463 enhanced status code where servers
// put it: immediately after the three digit reply code. Matching it there
// rather than anywhere in the reply avoids picking up IP addresses and
// diagnostic identifiers, which SMTP replies are full of.
var enhancedCodePattern = regexp.MustCompile(`^[245][0-9]{2}[ -]\s*([245]\.[0-9]{1,3}\.[0-9]{1,3})\b`)

// EnhancedCode returns the RFC 3463 enhanced status code the server sent, for
// example "5.7.1", or an empty string if it sent none. Servers are not obliged
// to send one and several large providers do not, so callers must handle the
// empty case rather than treating it as a classification.
//
// The first subcode is the class (2 success, 4 transient, 5 permanent) and the
// rest identify the condition: 5.1.x concerns the recipient address, while
// 5.7.x is a policy or security rejection that says nothing about whether the
// mailbox exists.
func (e *LookupError) EnhancedCode() string {
	if e == nil {
		return ""
	}
	return e.enhanced
}

// newLookupError creates a new LookupError reference and returns it
func newLookupError(message, details string) *LookupError {
	return &LookupError{Message: message, Details: details}
}

// withCause records the error this LookupError was derived from, so that
// errors.Is and errors.As can see through to it.
func (e *LookupError) withCause(cause error) *LookupError {
	if e == nil {
		return nil
	}
	e.cause = cause
	return e
}

// Unwrap returns the error this LookupError was derived from, if any.
func (e *LookupError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.cause
}

func (e *LookupError) Error() string {
	return fmt.Sprintf("%s : %s", e.Message, e.Details)
}

// ParseSMTPError receives an MX Servers response message
// and generates the corresponding MX error.
//
// A non-nil err always yields a non-nil result: an error we cannot classify is
// reported verbatim rather than discarded. Returning nil here would hand the
// caller a non-nil error interface wrapping a nil *LookupError, whose Error
// method then panics.
func ParseSMTPError(err error) *LookupError {
	if err == nil {
		return nil
	}
	errStr := err.Error()
	le := parseSMTPError(err)
	if le == nil {
		le = newLookupError(errStr, errStr)
	}
	if m := enhancedCodePattern.FindStringSubmatch(errStr); len(m) > 1 {
		le.enhanced = m[1]
	}
	return le.withCause(err)
}

func parseSMTPError(err error) *LookupError {
	errStr := err.Error()

	// Verify the length of the error before reading nil indexes
	if len(errStr) < 3 {
		return parseBasicErr(err)
	}

	// Strips out the status code string and converts to an integer for parsing
	status, convErr := strconv.Atoi(string([]rune(errStr)[0:3]))
	if convErr != nil {
		return parseBasicErr(err)
	}

	// If the status code is above 400 there was an error and we should return it
	if status > 400 {
		// Don't return an error if the error contains anything about the address
		// being undeliverable
		if insContains(errStr,
			"undeliverable",
			"does not exist",
			"may not exist",
			"user unknown",
			"user not found",
			"invalid address",
			"recipient invalid",
			"recipient rejected",
			"address rejected",
			"no mailbox") {
			return newLookupError(ErrServerUnavailable, errStr)
		}

		switch status {
		case 421:
			return newLookupError(ErrTryAgainLater, errStr)
		case 450:
			return newLookupError(ErrMailboxBusy, errStr)
		case 451:
			return newLookupError(ErrExceededMessagingLimits, errStr)
		case 452:
			if insContains(errStr,
				"full",
				"space",
				"over quota",
				"insufficient",
			) {
				return newLookupError(ErrFullInbox, errStr)
			}
			return newLookupError(ErrTooManyRCPT, errStr)
		case 503:
			return newLookupError(ErrNeedMAILBeforeRCPT, errStr)
		case 550: // 550 is Mailbox Unavailable - usually undeliverable, ref: https://blog.mailtrap.io/550-5-1-1-rejected-fix/
			if insContains(errStr,
				"spamhaus",
				"proofpoint",
				"cloudmark",
				"banned",
				"blacklisted",
				"blocked",
				"block list",
				"denied") {
				return newLookupError(ErrBlocked, errStr)
			}
			return newLookupError(ErrServerUnavailable, errStr)
		case 551:
			return newLookupError(ErrRCPTHasMoved, errStr)
		case 552:
			return newLookupError(ErrFullInbox, errStr)
		case 553:
			return newLookupError(ErrNoRelay, errStr)
		case 554:
			return newLookupError(ErrNotAllowed, errStr)
		default:
			return parseBasicErr(err)
		}
	}
	return nil
}

// parseBasicErr parses a basic MX record response and returns
// a more understandable LookupError
func parseBasicErr(err error) *LookupError {
	errStr := err.Error()

	// Return a more understandable error
	switch {
	case insContains(errStr,
		"spamhaus",
		"proofpoint",
		"cloudmark",
		"banned",
		"blocked",
		"denied"):
		return newLookupError(ErrBlocked, errStr)
	case insContains(errStr, "timeout"):
		return newLookupError(ErrTimeout, errStr)
	case insContains(errStr, "no such host"):
		return newLookupError(ErrNoSuchHost, errStr)
	case insContains(errStr, "unavailable"):
		return newLookupError(ErrServerUnavailable, errStr)
	default:
		return newLookupError(errStr, errStr)
	}
}

// insContains returns true if any of the substrings
// are found in the passed string. This method of checking
// contains is case insensitive
func insContains(str string, subStrs ...string) bool {
	for _, subStr := range subStrs {
		if strings.Contains(strings.ToLower(str),
			strings.ToLower(subStr)) {
			return true
		}
	}
	return false
}
