package emailverifier

import (
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCheckSMTPUnSupportedVendor(t *testing.T) {
	err := verifier.EnableAPIVerifier("unsupported_vendor")
	assert.Error(t, err)
}

func TestCheckSMTPOK_ByApi(t *testing.T) {
	cases := []struct {
		name     string
		domain   string
		username string
		expected *SMTP
	}{
		{
			name:     "yahoo exists",
			domain:   "yahoo.com",
			username: "someone",
			expected: &SMTP{
				HostExists:  true,
				Deliverable: true,
			},
		},
		{
			name:     "myyahoo exists",
			domain:   "myyahoo.com",
			username: "someone",
			expected: &SMTP{
				HostExists:  true,
				Deliverable: true,
			},
		},
		{
			name:     "yahoo no exists",
			domain:   "yahoo.com",
			username: "123",
			expected: &SMTP{
				HostExists:  true,
				Deliverable: false,
			},
		},
		{
			name:     "myyahoo no exists",
			domain:   "myyahoo.com",
			username: "123",
			expected: &SMTP{
				HostExists:  true,
				Deliverable: false,
			},
		},
	}
	_ = verifier.EnableAPIVerifier(YAHOO)
	defer verifier.DisableAPIVerifier(YAHOO)
	for _, c := range cases {
		test := c
		t.Run(test.name, func(tt *testing.T) {
			smtp, err := verifier.CheckSMTP(test.domain, test.username)
			assert.NoError(t, err)
			assert.Equal(t, test.expected, smtp)
		})
	}
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
	ret, _, err := newSMTPClient(domain, "", timeout, timeout)
	assert.NotNil(t, ret)
	assert.Nil(t, err)
}

func TestNewSMTPClientFailed_WithInvalidProxy(t *testing.T) {
	domain := "gmail.com"
	proxyURI := "socks5://user:password@127.0.0.1:1080?timeout=5s"
	timeout := 5 * time.Second
	ret, _, err := newSMTPClient(domain, proxyURI, timeout, timeout)
	assert.Nil(t, ret)
	assert.Error(t, err, syscall.ECONNREFUSED)
}

func TestNewSMTPClientFailed(t *testing.T) {
	domain := "zzzz171777.com"
	timeout := 5 * time.Second
	ret, _, err := newSMTPClient(domain, "", timeout, timeout)
	assert.Nil(t, ret)
	assert.Error(t, err)
}

func TestDialSMTPFailed_NoPortIsConfigured(t *testing.T) {
	disposableDomain := "zzzz1717.com"
	timeout := 5 * time.Second
	ret, err := dialSMTP(disposableDomain, "", timeout, timeout)
	assert.Nil(t, ret)
	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "missing port"))
}

func TestDialSMTPFailed_NoSuchHost(t *testing.T) {
	disposableDomain := "zzzzyyyyaaa123.com:25"
	timeout := 5 * time.Second
	ret, err := dialSMTP(disposableDomain, "", timeout, timeout)
	assert.Nil(t, ret)
	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "no such host"))
}
