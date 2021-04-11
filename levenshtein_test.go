package emailverifier

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLevenshteinDistanceOK1(t *testing.T) {
	s1, s2 := "gmail.com", "gmaii.com"
	assert.Equal(t, levenshteinDistance(s1, s2), 1)
}

func TestLevenshteinDistanceOK2(t *testing.T) {
	s1, s2 := "gmail.com", "gmai.com"
	assert.Equal(t, levenshteinDistance(s1, s2), 1)
}

func TestLevenshteinDistanceOK3(t *testing.T) {
	s1, s2 := "", "abcde"
	assert.Equal(t, levenshteinDistance(s1, s2), 5)
}

func TestLevenshteinDistanceOK4(t *testing.T) {
	s1, s2 := "abcde", "abcde"
	assert.Equal(t, levenshteinDistance(s1, s2), 0)
}

func TestLevenshteinDistanceOK5(t *testing.T) {
	s1, s2 := "distance", "difference"
	assert.Equal(t, levenshteinDistance(s1, s2), 5)
}
