package emailverifier

import (
	"fmt"
	"testing"
)

func Test_levenshteinDistance(t *testing.T) {
	tests := []struct {
		str1 string
		str2 string
		want int
	}{
		{"", "abcde", 5},
		{"abcde", "", 5},
		{"abcde", "abcde", 0},
		{"ab", "aa", 1},
		{"ab", "ba", 2},
		{"ab", "aaa", 2},
		{"gmail.com", "gmaii.com", 1},
		{"gmail.com", "gmai.com", 1},
		{"bbb", "a", 3},
		{"kitten", "sitting", 3},
		{"distance", "difference", 5},
		{"a cat", "an abct", 4},
	}
	for i, tt := range tests {
		t.Run(fmt.Sprintf("test_case_%d", i+1), func(t *testing.T) {
			if got := levenshteinDistance(tt.str1, tt.str2); got != tt.want {
				t.Errorf("levenshteinDistance() = %v, want %v", got, tt.want)
			}
		})
	}
}
