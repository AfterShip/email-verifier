package emailverifier

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckMxWithCacheOK(t *testing.T) {
	domain := "github.com"

	mx, err := verifier.CheckMX(domain, true)
	assert.NoError(t, err)
	assert.True(t, mx.HasMXRecord)
}

func TestCheckNoMxWithCacheOK(t *testing.T) {
	domain := "githubexists.com"

	mx, err := verifier.CheckMX(domain, true)
	assert.Nil(t, mx)
	assert.Error(t, err, ErrNoSuchHost)
}
