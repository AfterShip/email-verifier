package emailverifier

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckEmailOK_SMTPHostNotExists(t *testing.T) {
	var (
		// trueVal  = true
		username = "email_username"
		domain   = "domainnotexists.com"
		address  = username + "@" + domain
		email    = address
	)

	ret, err := verifier.Verify(email)
	expected := Result{
		Email: email,
		Syntax: &Syntax{
			Username: username,
			Domain:   domain,
			Valid:    true,
		},
		HasMxRecords: false,
		Disposable:   false,
		RoleAccount:  false,
		Reachable:    reachableUnknown,
		Free:         false,
		SMTP:         nil,
	}
	assert.Error(t, err, ErrNoSuchHost)
	assert.Equal(t, &expected, ret)
}

func TestCheckEmailOK_SMTPHostExists_NotCatchAll(t *testing.T) {
	var (
		// trueVal  = true
		username = "email_username"
		domain   = "github.com"
		address  = username + "@" + domain
		email    = address
	)

	ret, err := verifier.Verify(email)
	expected := Result{
		Email: email,
		Syntax: &Syntax{
			Username: username,
			Domain:   domain,
			Valid:    true,
		},
		HasMxRecords: true,
		Reachable:    reachableUnknown,
		Disposable:   false,
		RoleAccount:  false,
		Free:         false,
		SMTP: &SMTP{
			HostExists:  true,
			FullInbox:   false,
			CatchAll:    true,
			Deliverable: false,
			Disabled:    false,
		},
	}
	assert.Nil(t, err)
	assert.Equal(t, &expected, ret)
}

func TestCheckEmailOK_SMTPHostExists_CatchAll(t *testing.T) {
	var (
		// trueVal  = true
		username = "email_username"
		domain   = "yahoo.com"
		address  = username + "@" + domain
		email    = address
	)

	ret, err := verifier.Verify(email)
	expected := Result{
		Email: email,
		Syntax: &Syntax{
			Username: username,
			Domain:   domain,
			Valid:    true,
		},
		HasMxRecords: true,
		Reachable:    reachableUnknown,
		Disposable:   false,
		RoleAccount:  false,
		Free:         true,
		SMTP: &SMTP{
			HostExists:  true,
			FullInbox:   false,
			CatchAll:    true,
			Deliverable: false,
			Disabled:    false,
		},
	}
	assert.Nil(t, err)
	assert.Equal(t, &expected, ret)
}

func TestCheckEmailOK_SMTPHostExists_FreeDomain(t *testing.T) {
	var (
		// trueVal  = true
		username = "email_username"
		domain   = "yahoo.com"
		address  = username + "@" + domain
		email    = address
	)

	ret, err := verifier.Verify(email)
	expected := Result{
		Email: email,
		Syntax: &Syntax{
			Username: username,
			Domain:   domain,
			Valid:    true,
		},
		HasMxRecords: true,
		Reachable:    reachableUnknown,
		Disposable:   false,
		RoleAccount:  false,
		Free:         true,
		SMTP: &SMTP{
			HostExists:  true,
			FullInbox:   false,
			CatchAll:    true,
			Deliverable: false,
			Disabled:    false,
		},
	}
	assert.Nil(t, err)
	assert.Equal(t, &expected, ret)
}

func TestCheckEmail_ErrorSyntax(t *testing.T) {
	var (
		// trueVal  = true
		username = ""
		domain   = "yahoo.com"
		address  = username + "@" + domain
		email    = address
	)

	ret, err := verifier.Verify(email)
	expected := Result{
		Email: email,
		Syntax: &Syntax{
			Username: username,
			Domain:   "",
			Valid:    false,
		},
		HasMxRecords: false,
		Reachable:    reachableUnknown,
		Disposable:   false,
		RoleAccount:  false,
		Free:         false,
		SMTP:         nil,
	}
	assert.Nil(t, err)
	assert.Equal(t, &expected, ret)
}

func TestCheckEmail_Disposable(t *testing.T) {
	var (
		// trueVal  = true
		username = "exampleuser"
		domain   = "zzjbfwqi.shop"
		address  = username + "@" + domain
		email    = address
	)

	ret, err := verifier.Verify(email)
	expected := Result{
		Email: email,
		Syntax: &Syntax{
			Username: username,
			Domain:   domain,
			Valid:    true,
		},
		HasMxRecords: false,
		Reachable:    reachableUnknown,
		Disposable:   true,
		RoleAccount:  false,
		Free:         false,
		SMTP:         nil,
	}
	assert.Nil(t, err)
	assert.Equal(t, &expected, ret)
}

func TestCheckEmail_RoleAccount(t *testing.T) {
	var (
		// trueVal  = true
		username = "admin"
		domain   = "github.com"
		address  = username + "@" + domain
		email    = address
	)

	ret, err := verifier.Verify(email)
	expected := Result{
		Email: email,
		Syntax: &Syntax{
			Username: username,
			Domain:   domain,
			Valid:    true,
		},
		HasMxRecords: true,
		Reachable:    reachableUnknown,
		Disposable:   false,
		RoleAccount:  true,
		Free:         false,
		SMTP: &SMTP{
			HostExists:  true,
			FullInbox:   false,
			CatchAll:    true,
			Deliverable: false,
			Disabled:    false,
		},
	}
	assert.Nil(t, err)
	assert.Equal(t, &expected, ret)
}

func TestCheckEmail_DisabledSMTPCheck(t *testing.T) {
	var (
		// trueVal  = true
		username = "email_username"
		domain   = "randomain.com"
		address  = username + "@" + domain
		email    = address
	)

	verifier.DisableSMTPCheck()
	ret, err := verifier.Verify(email)
	expected := Result{
		Email: email,
		Syntax: &Syntax{
			Username: username,
			Domain:   domain,
			Valid:    true,
		},
		HasMxRecords: true,
		Disposable:   false,
		RoleAccount:  false,
		Reachable:    reachableUnknown,
		Free:         false,
		SMTP:         nil,
	}
	verifier.EnableSMTPCheck()
	assert.NoError(t, err)
	assert.Equal(t, &expected, ret)
}

func TestNewVerifierOK_AutoUpdateDisposable(t *testing.T) {
	verifier.EnableAutoUpdateDisposable()
}

func TestNewVerifierOK_EnableAutoUpdateDisposable(t *testing.T) {
	verifier.EnableAutoUpdateDisposable()
}

func TestNewVerifierOK_AutoUpdateDisposableDuplicate(t *testing.T) {
	verifier.DisableAutoUpdateDisposable()

	verifier.EnableAutoUpdateDisposable()
	verifier.DisableAutoUpdateDisposable()

	verifier.EnableAutoUpdateDisposable()
	verifier.DisableAutoUpdateDisposable()
	verifier.EnableAutoUpdateDisposable()
}

func TestStopCurrentSchedule_ScheduleIsNil(t *testing.T) {
	verifier.schedule = nil
	verifier.stopCurrentSchedule()
}

func TestStopCurrentScheduleOK(t *testing.T) {
	verifier.EnableAutoUpdateDisposable()
	verifier.stopCurrentSchedule()
}
