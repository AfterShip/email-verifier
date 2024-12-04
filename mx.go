package emailverifier

import (
	"net"
)

// Mx is detail about the Mx host
type Mx struct {
	HasMXRecord bool      // whether has 1 or more MX record
	Records     []*net.MX // represent DNS MX records
}

// CheckMX will return the DNS MX records for the given domain name sorted by preference.
func (v *Verifier) CheckMX(domain string, useCache bool) (*Mx, error) {
	domain = domainToASCII(domain)

	var mx []*net.MX
	var err error
	
	lookup := net.LookupMX
	if useCache {
		lookup = v.mxCache.Get
	}
	mx, err = lookup(domain)

	if err != nil && len(mx) == 0 {
		return nil, err
	}
	return &Mx{
		HasMXRecord: len(mx) > 0,
		Records:     mx,
	}, nil
}
