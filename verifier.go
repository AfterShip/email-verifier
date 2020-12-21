package emailverifier

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

var (
	_, b, _, _ = runtime.Caller(0)
	basePath   = filepath.Dir(b)
)

// Verifier is an email verifier. Create one by calling NewVerifier
type Verifier struct {
	smtpCheckEnabled bool      // SMTP check enabled or disabled (disabled by default)
	fromEmail        string    // name to use in the `EHLO:` SMTP command, defaults to "user@example.org"
	helloName        string    // email to use in the `MAIL FROM:` SMTP command. defaults to `localhost`
	schedule         *schedule // schedule represents a job schedule
}

// Result is the result of Email Verification
type Result struct {
	Email        string  `json:"email"`          // passed email address
	Disposable   bool    `json:"disposable"`     // is this a DEA (disposable email address)
	Reachable    string  `json:"reachable"`      // an enumeration to describe whether the recipient address is real
	RoleAccount  bool    `json:"role_account"`   // is account a role-based account
	Free         bool    `json:"free"`           // is domain a free email domain
	Syntax       *Syntax `json:"syntax"`         // details about the email address syntax
	HasMxRecords bool    `json:"has_mx_records"` // whether or not MX-Records for the domain
	SMTP         *SMTP   `json:"smtp"`           // details about the SMTP response of the email
}

// NewVerifier creates a new email verifier
func NewVerifier() *Verifier {
	loadDisposableDomains()
	loadFreeDomains()
	loadRoleAccounts()

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

	// If domain is disposable, do not check mx and smtp. Because domain probably doesn't exist.
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

	return &ret, nil
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

// FromEmail set the emails to use in the `MAIL FROM:` smtp command
func (v *Verifier) FromEmail(email string) *Verifier {
	v.fromEmail = email
	return v
}

// HelloName set the name to use in the `EHLO:` SMTP command
func (v *Verifier) HelloName(domain string) *Verifier {
	v.helloName = domain
	return v
}

// loadFreeDomains loads free_domain data
func loadFreeDomains() {
	if len(freeDomains) > 0 {
		return
	}

	freeDomainFile, err := os.Open(basePath + freeDomainDataPath)
	if err != nil {
		panic(fmt.Sprintf("open free domains' data file fail: %v ", err))
	}

	scanner := bufio.NewScanner(freeDomainFile)
	scanner.Split(bufio.ScanLines)

	freeDomains = make(map[string]bool)
	for scanner.Scan() {
		freeDomains[scanner.Text()] = true
	}

	err = freeDomainFile.Close()
	if err != nil {
		panic(fmt.Sprintf("close free domains' data file fail: %v ", err))
	}
}

// loadDisposableDomains loads disposable_domain data
func loadDisposableDomains() {
	if disposableDomainsLoaded {
		return
	}

	disposableDomainFile, err := os.Open(basePath + disposableDomainDataPath)
	if err != nil {
		panic(fmt.Sprintf("open disposable domains' data file fail: %v ", err))
	}

	scanner := bufio.NewScanner(disposableDomainFile)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		disposableDomains.Store(scanner.Text(), struct{}{})
	}

	err = disposableDomainFile.Close()
	if err != nil {
		panic(fmt.Sprintf("close disposable domains' data file fail: %v ", err))
	}
	disposableDomainsLoaded = true
}

// loadRoleAccounts loads role_account data
func loadRoleAccounts() {
	if len(roleAccounts) > 0 {
		return
	}

	roleAccountFile, err := os.Open(basePath + roleAccountDataPath)
	if err != nil {
		panic(fmt.Sprintf("open role accounts' data file fail: %v ", err))
	}

	scanner := bufio.NewScanner(roleAccountFile)
	scanner.Split(bufio.ScanLines)

	roleAccounts = make(map[string]bool)
	for scanner.Scan() {
		roleAccounts[scanner.Text()] = true
	}

	err = roleAccountFile.Close()
	if err != nil {
		panic(fmt.Sprintf("close role accounts' data file fail: %v ", err))
	}
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
