package emailverifier

import (
	"context"
	"errors"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckMxOK(t *testing.T) {
	domain := "github.com"

	mx, err := verifier.CheckMX(domain)
	assert.NoError(t, err)
	assert.True(t, mx.HasMXRecord)
}

func TestCheckNoMxOK(t *testing.T) {
	domain := "githubexists.com"

	mx, err := verifier.CheckMX(domain)
	assert.Nil(t, mx)
	assert.Error(t, err, ErrNoSuchHost)
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
