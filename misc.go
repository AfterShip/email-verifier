package emailverifier

import (
	"strings"
	"sync"
)

var (
	disposableDomains       sync.Map        // map to store disposable domains data
	disposableDomainsLoaded bool            //  whether disposableDomains is loaded or not
	freeDomains             map[string]bool // map to store free domains data
	roleAccounts            map[string]bool // map to store role-based accounts data
)

// IsRoleAccount checks if username is a role-based account
func (v *Verifier) IsRoleAccount(username string) bool {
	return roleAccounts[strings.ToLower(username)]
}

// IsFreeDomain checks if domain is a free domain
func (v *Verifier) IsFreeDomain(domain string) bool {
	return freeDomains[domain]
}

// IsDisposable checks if domain is a disposable domain
func (v *Verifier) IsDisposable(domain string) bool {
	domain = domainToASCII(domain)
	d := parsedDomain(domain)
	_, found := disposableDomains.Load(d)
	return found
}
