package emailverifier

import (
	"strings"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
	domain := "yahoo.com"

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
	domain := "yahoo.com"
	ret, err := newSMTPClient(domain, "")
	assert.NotNil(t, ret)
	assert.Nil(t, err)
}

func TestNewSMTPClientFailed_WithInvalidProxy(t *testing.T) {
	domain := "yahoo.com"
	proxyURI := "socks5://user:password@127.0.0.1:1080?timeout=5s"
	ret, err := newSMTPClient(domain, proxyURI)
	assert.Nil(t, ret)
	assert.Error(t, err, syscall.ECONNREFUSED)
}

func TestNewSMTPClientFailed(t *testing.T) {
	disposableDomain := "zzzz1717.com"
	ret, err := newSMTPClient(disposableDomain, "")
	assert.Nil(t, ret)
	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "no such host"))
}

func TestDialSMTPFailed_NoPortIsConfigured(t *testing.T) {
	disposableDomain := "zzzz1717.com"
	ret, err := dialSMTP(disposableDomain, "")
	assert.Nil(t, ret)
	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "missing port"))
}

func TestDialSMTPFailed_NoSuchHost(t *testing.T) {
	disposableDomain := "zzzzyyyyaaa123.com:25"
	ret, err := dialSMTP(disposableDomain, "")
	assert.Nil(t, ret)
	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "no such host"))
}
