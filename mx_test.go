package emailverifier

import (
	"context"
	"errors"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckMxOK(t *testing.T) {
	domain := "github.com"

	mx, err := verifier.CheckMX(domain)
	require.NoError(t, err)
	assert.True(t, mx.HasMXRecord)
}

func TestCheckNoMxOK(t *testing.T) {
	domain := "githubexists.com"

	mx, err := verifier.CheckMX(domain)
	assert.Nil(t, mx)

	// CheckMX surfaces the resolver error as-is; it is Verify that maps it to
	// ErrNoSuchHost. Assert the DNS semantics rather than the message text,
	// which carries the resolver address and so varies by environment.
	var dnsErr *net.DNSError
	require.ErrorAs(t, err, &dnsErr)
	assert.True(t, dnsErr.IsNotFound, "expected NXDOMAIN, got %v", err)
}

func TestCheckMx_WithCustomResolver(t *testing.T) {
	domain := "github.com"
	wantErr := errors.New("custom resolver dial invoked")

	customResolver := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			return nil, wantErr
		},
	}

	v := NewVerifier().Resolver(customResolver)
	mx, err := v.CheckMX(domain)
	assert.Nil(t, mx)
	assert.ErrorContains(t, err, wantErr.Error())
}
