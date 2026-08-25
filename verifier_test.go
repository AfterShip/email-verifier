package emailverifier

import (
	"net"
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
		Syntax: Syntax{
			Username: username,
			Domain:   domain,
			Valid:    true,
		},
		HasMxRecords: false,
		Disposable:   false,
		RoleAccount:  false,
		Reachable:    reachableNo,
		Free:         false,
		SMTP:         nil,
	}
	assert.ErrorContains(t, err, ErrNoSuchHost)
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
		Syntax: Syntax{
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

func TestCheckEmailOK_SMTPHostExists_FreeDomain(t *testing.T) {
	var (
		// trueVal  = true
		username = "email_username"
		domain   = "gmail.com"
		address  = username + "@" + domain
		email    = address
	)

	ret, err := verifier.Verify(email)
	expected := Result{
		Email: email,
		Syntax: Syntax{
			Username: username,
			Domain:   domain,
			Valid:    true,
		},
		HasMxRecords: true,
		Reachable:    reachableNo,
		Disposable:   false,
		RoleAccount:  false,
		Free:         true,
		SMTP: &SMTP{
			HostExists:  true,
			FullInbox:   false,
			CatchAll:    false,
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
		Syntax: Syntax{
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
		Syntax: Syntax{
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

func TestCheckEmail_Disposable_override(t *testing.T) {
	var (
		username = "exampleuser"
		domain   = "iamdisposableemail.test"
		address  = username + "@" + domain
		email    = address
	)

	verifier := NewVerifier().EnableSMTPCheck().AddDisposableDomains([]string{"iamdisposableemail.test"})
	ret, err := verifier.Verify(email)
	expected := Result{
		Email: email,
		Syntax: Syntax{
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
		Syntax: Syntax{
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
		Syntax: Syntax{
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

func TestCheckEmail_EnableDomainSuggest(t *testing.T) {
	var (
		// trueVal  = true
		username = "email_username"
		domain   = "hotmail.com"
		address  = username + "@" + domain
		email    = address
	)

	ret, _ := verifier.Verify(email)

	assert.Empty(t, ret.Suggestion)
}

func TestCheckEmail_EnableDomainSuggest_Gmail(t *testing.T) {
	var (
		// trueVal  = true
		username = "email_username"
		domain   = "gmai.com"
		address  = username + "@" + domain
		email    = address
	)

	ret, _ := verifier.EnableDomainSuggest().Verify(email)

	assert.Equal(t, "gmail.com", ret.Suggestion)
}

func TestNewVerifierOK_DefaultResolver(t *testing.T) {
	v := NewVerifier()
	assert.Nil(t, v.resolver)
	assert.Same(t, net.DefaultResolver, v.dnsResolver())
}

func TestVerifier_Resolver(t *testing.T) {
	customResolver := &net.Resolver{PreferGo: true}

	v := NewVerifier().Resolver(customResolver)

	assert.Same(t, customResolver, v.resolver)
	assert.Same(t, customResolver, v.dnsResolver())
}

func TestVerifier_Resolver_NilRestoresDefault(t *testing.T) {
	customResolver := &net.Resolver{PreferGo: true}

	v := NewVerifier().Resolver(customResolver)
	assert.Same(t, customResolver, v.dnsResolver())

	v.Resolver(nil)

	assert.Same(t, net.DefaultResolver, v.dnsResolver())
}

// net.DefaultResolver must be read at lookup time, not captured when the
// Verifier is built. Callers who replace the global -- the only way to redirect
// lookups before Resolver() existed -- would otherwise be silently ignored,
// since NewVerifier is commonly called during package variable initialisation,
// before any init() that installs the replacement.
func TestVerifier_DefaultResolverReadAtCallTime(t *testing.T) {
	original := net.DefaultResolver
	t.Cleanup(func() { net.DefaultResolver = original })

	v := NewVerifier()

	replacement := &net.Resolver{PreferGo: true}
	net.DefaultResolver = replacement

	assert.Same(t, replacement, v.dnsResolver())
}
