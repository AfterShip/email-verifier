package emailverifier

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var verifier = NewVerifier().EnableSMTPCheck()

func TestIsFreeDomain_True(t *testing.T) {
	domain := "yahoo.com"

	isFreeDomain := verifier.IsFreeDomain(domain)
	assert.True(t, isFreeDomain)
}

func TestCheckNotFreeDomain_False(t *testing.T) {
	domain := "github.com"

	isFreeDomain := verifier.IsFreeDomain(domain)
	assert.False(t, isFreeDomain)
}

func TestIsDisposableDomain_True(t *testing.T) {
	domain := "dbbd8.club"

	isDisposable := verifier.IsDisposable(domain)
	assert.True(t, isDisposable)
}

func TestIsDisposableDomain_False(t *testing.T) {
	domain := "gmail.com"

	isDisposable := verifier.IsDisposable(domain)
	assert.False(t, isDisposable)
}

func TestIsRoleAccount_True(t *testing.T) {
	username := "administrator"

	isRoleAccount := verifier.IsRoleAccount(username)
	assert.True(t, isRoleAccount)
}

func TestIsRoleAccount_False(t *testing.T) {
	username := "normal_user"

	isRoleAccount := verifier.IsRoleAccount(username)
	assert.False(t, isRoleAccount)
}
