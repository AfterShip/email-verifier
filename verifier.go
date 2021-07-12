package emailverifier

import (
	"time"
)

// Verifier is an email verifier. Create one by calling NewVerifier
type Verifier struct {
	smtpCheckEnabled     bool      // SMTP check enabled or disabled (disabled by default)
	domainSuggestEnabled bool      // whether suggest a most similar correct domain or not (disabled by default)
	gravatarCheckEnabled bool      // gravatar check enabled or disabled (disabled by default)
	fromEmail            string    // name to use in the `EHLO:` SMTP command, defaults to "user@example.org"
	helloName            string    // email to use in the `MAIL FROM:` SMTP command. defaults to `localhost`
	schedule             *schedule // schedule represents a job schedule
}

// Result is the result of Email Verification
type Result struct {
	Email        string    `json:"email"`          // passed email address
	Reachable    string    `json:"reachable"`      // an enumeration to describe whether the recipient address is real
	Syntax       Syntax    `json:"syntax"`         // details about the email address syntax
	SMTP         *SMTP     `json:"smtp"`           // details about the SMTP response of the email
	Gravatar     *Gravatar `json:"gravatar"`       // whether or not have gravatar for the email
	Suggestion   string    `json:"suggestion"`     // domain suggestion when domain is misspelled
	Disposable   bool      `json:"disposable"`     // is this a DEA (disposable email address)
	RoleAccount  bool      `json:"role_account"`   // is account a role-based account
	Free         bool      `json:"free"`           // is domain a free email domain
	HasMxRecords bool      `json:"has_mx_records"` // whether or not MX-Records for the domain
}

// additional list of disposable domains set via users of this library
var additionalDisposableDomains map[string]bool = map[string]bool{}

// init loads disposable_domain meta data to disposableSyncDomains which are safe for concurrent use
func init() {
	for d := range disposableDomains {
		disposableSyncDomains.Store(d, struct{}{})
	}
}

// NewVerifier creates a new email verifier
func NewVerifier() *Verifier {
	return &Verifier{
		fromEmail: defaultFromEmail,
		helloName: defaultHelloName,
	}
}

// Verify performs address, misc, mx and smtp checks
func (v *Verifier) Verify(email string) (*Result, error) {

	ret := Result{
		Email:     email,
		Reachable: reachableUnknown,
	}

	syntax := v.ParseAddress(email)
	ret.Syntax = syntax
	if !syntax.Valid {
		return &ret, nil
	}

	ret.Free = v.IsFreeDomain(syntax.Domain)
	ret.RoleAccount = v.IsRoleAccount(syntax.Username)
	ret.Disposable = v.IsDisposable(syntax.Domain)

	// If the domain name is disposable, mx and smtp are not checked.
	if ret.Disposable {
		return &ret, nil
	}

	mx, err := v.CheckMX(syntax.Domain)
	if err != nil {
		return &ret, err
	}
	ret.HasMxRecords = mx.HasMXRecord

	smtp, err := v.CheckSMTP(syntax.Domain, syntax.Username)
	if err != nil {
		return &ret, err
	}
	ret.SMTP = smtp
	ret.Reachable = v.calculateReachable(smtp)

	if v.gravatarCheckEnabled {
		gravatar, err := v.CheckGravatar(email)
		if err != nil {
			return &ret, err
		}
		ret.Gravatar = gravatar
	}

	if v.domainSuggestEnabled {
		ret.Suggestion = v.SuggestDomain(syntax.Domain)
	}

	return &ret, nil
}

// AddDisposableDomains adds additional domains as disposable domains.
func (v *Verifier) AddDisposableDomains(domains []string) *Verifier {
	for _, d := range domains {
		additionalDisposableDomains[d] = true
		disposableSyncDomains.Store(d, struct{}{})
	}
	return v
}

// EnableGravatarCheck enables check gravatar,
// we don't check gravatar by default
func (v *Verifier) EnableGravatarCheck() *Verifier {
	v.gravatarCheckEnabled = true
	return v
}

// DisableGravatarCheck disables check gravatar,
func (v *Verifier) DisableGravatarCheck() *Verifier {
	v.gravatarCheckEnabled = false
	return v
}

// EnableSMTPCheck enables check email by smtp,
// for most ISPs block outgoing SMTP requests through port 25, to prevent spam,
// we don't check smtp by default
func (v *Verifier) EnableSMTPCheck() *Verifier {
	v.smtpCheckEnabled = true
	return v
}

// DisableSMTPCheck disables check email by smtp
func (v *Verifier) DisableSMTPCheck() *Verifier {
	v.smtpCheckEnabled = false
	return v
}

// EnableDomainSuggest will suggest a most similar correct domain when domain misspelled
func (v *Verifier) EnableDomainSuggest() *Verifier {
	v.domainSuggestEnabled = true
	return v
}

// DisableDomainSuggest will not suggest anything
func (v *Verifier) DisableDomainSuggest() *Verifier {
	v.domainSuggestEnabled = false
	return v
}

// EnableAutoUpdateDisposable enables update disposable domains automatically
func (v *Verifier) EnableAutoUpdateDisposable() *Verifier {
	v.stopCurrentSchedule()

	// update disposable domains records daily
	v.schedule = newSchedule(24*time.Hour, updateDisposableDomains, disposableDataURL)
	v.schedule.start()
	return v
}

// DisableAutoUpdateDisposable stops previously started schedule job
func (v *Verifier) DisableAutoUpdateDisposable() *Verifier {
	v.stopCurrentSchedule()
	return v

}

// FromEmail sets the emails to use in the `MAIL FROM:` smtp command
func (v *Verifier) FromEmail(email string) *Verifier {
	v.fromEmail = email
	return v
}

// HelloName sets the name to use in the `EHLO:` SMTP command
func (v *Verifier) HelloName(domain string) *Verifier {
	v.helloName = domain
	return v
}

func (v *Verifier) calculateReachable(s *SMTP) string {
	if !v.smtpCheckEnabled {
		return reachableUnknown
	}
	if s.Deliverable {
		return reachableYes
	}
	if s.CatchAll {
		return reachableUnknown
	}
	return reachableNo
}

// stopCurrentSchedule stops current running schedule (if exists)
func (v *Verifier) stopCurrentSchedule() {
	if v.schedule != nil {
		v.schedule.stop()
	}
}
