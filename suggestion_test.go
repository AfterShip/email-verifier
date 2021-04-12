package emailverifier

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSuggestDomainOK_HitExactDomain(t *testing.T) {
	domain := "gmail.com"

	ret := verifier.SuggestDomain(domain)
	assert.Equal(t, "", ret)
}

func TestSuggestDomainOK_NullString(t *testing.T) {
	domain := ""

	ret := verifier.SuggestDomain(domain)
	assert.Equal(t, "", ret)
}

func TestSuggestDomainOK_SimilarDomain1(t *testing.T) {
	domain := "gmaii.com"

	ret := verifier.SuggestDomain(domain)
	assert.Equal(t, "gmail.com", ret)
}

func TestSuggestDomainOK_SimilarDomain2(t *testing.T) {
	domain := "gmai.com"

	ret := verifier.SuggestDomain(domain)
	assert.Equal(t, "gmail.com", ret)
}

func TestSuggestDomainOK_TLD(t *testing.T) {
	domain := "gmail.edd"

	ret := verifier.SuggestDomain(domain)
	assert.Equal(t, "gmail.edu", ret)
}

func TestSuggestDomainOK_SLD(t *testing.T) {
	domain := "homail.aftership"

	ret := verifier.SuggestDomain(domain)
	assert.Equal(t, "hotmail.aftership", ret)
}
