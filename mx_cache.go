package emailverifier

import (
	"net"
	"sync"
	"time"
)

// MXCache represents a thread-safe cache for MX records
type MXCache struct {
	sync.RWMutex
	records map[string]cacheEntry
	ttl     time.Duration
}

type cacheEntry struct {
	mxRecords []*net.MX
	expiry    time.Time
}

// NewMXCache creates and initializes a new MX cache and starts the cleanup goroutine
func NewMXCache(ttl time.Duration) *MXCache {
	cache := &MXCache{
		records: make(map[string]cacheEntry),
		ttl:     ttl,
	}

	// Start the cleanup process with a 15-minute interval
	cache.StartCleanup(15 * time.Minute)

	return cache
}

// Get retrieves MX records for a domain from cache or performs a lookup
func (c *MXCache) Get(domain string) ([]*net.MX, error) {
	asciiDomain := domainToASCII(domain)

	// Check cache first
	c.RLock()
	if entry, exists := c.records[asciiDomain]; exists {
		if time.Now().Before(entry.expiry) {
			c.RUnlock()
			return entry.mxRecords, nil
		}
	}
	c.RUnlock()

	// Perform actual lookup if not in cache or expired
	mxRecords, err := net.LookupMX(asciiDomain)
	if err != nil {
		return nil, err
	}

	// Update cache
	c.Lock()
	c.records[asciiDomain] = cacheEntry{
		mxRecords: mxRecords,
		expiry:    time.Now().Add(c.ttl),
	}
	c.Unlock()

	return mxRecords, nil
}

// StartCleanup initiates periodic cleanup of expired cache entries
func (c *MXCache) StartCleanup(interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			c.Lock()
			now := time.Now()
			for domain, entry := range c.records {
				if now.After(entry.expiry) {
					delete(c.records, domain)
				}
			}
			c.Unlock()
		}
	}()
}
