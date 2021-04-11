package emailverifier

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStringsSimilaritystr1Longer(t *testing.T) {
	s1, s2 := "Automizely", "AfterShip"
	assert.Greater(t, stringsSimilarity(s1, s2, 3), 0.5)
}

func TestStringsSimilaritystr2Longer(t *testing.T) {
	s2, s1 := "Automizely", "AfterShip"
	assert.Less(t, stringsSimilarity(s1, s2, 3), float32(0.8))
}

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
