package emailverifier

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDomainToASCII(t *testing.T) {
	domain := "testingΣ✪✯☭➳卐.org" //nolint:gosmopolitan // deliberate non-ASCII IDN input
	ret := domainToASCII(domain)
	expected := "xn--testing-0if2960fjccubz8h9z13a.org"
	assert.Equal(t, expected, ret)
}

func TestDomainToASCII_NormalDomain(t *testing.T) {
	domain := "testing.org"
	ret := domainToASCII(domain)
	assert.Equal(t, domain, ret)
}

func TestCallJobFuncWithParams_NoOutput(t *testing.T) {
	f := func(a string) { fmt.Println(a) }
	ret := callJobFuncWithParams(f, []interface{}{"testing"})
	assert.Nil(t, ret)
}

func TestCallJobFuncWithParams_WithOutput(t *testing.T) {
	f := func(a int) int { return a * (1 + a) }
	ret := callJobFuncWithParams(f, []interface{}{2})
	assert.Equal(t, int64(6), ret[0].Int())
}

func TestCallJobFuncForgetParams(t *testing.T) {
	f := func(a int) int { return a * (1 + a) }
	ret := callJobFuncWithParams(f, nil)
	assert.Nil(t, ret)
}

func TestCallJobFuncWithWrongFunc(t *testing.T) {
	f := 3
	ret := callJobFuncWithParams(f, nil)
	assert.Nil(t, ret)
}

func TestSplitDomainNoSLD(t *testing.T) {
	domain := "com"
	sld, tld := splitDomain(domain)
	assert.Empty(t, sld)
	assert.Equal(t, domain, tld)
}

func TestSplitDomainOK(t *testing.T) {
	domain := "aftership.com"
	sld, tld := splitDomain(domain)
	assert.Equal(t, "aftership", sld)
	assert.Equal(t, "com", tld)
}

func TestSplitDomainNilString(t *testing.T) {
	domain := ""
	sld, tld := splitDomain(domain)
	assert.Empty(t, sld)
	assert.Empty(t, tld)
}

func TestSplitDomainSubDomain(t *testing.T) {
	domain := "develop.aftership.com"
	sld, tld := splitDomain(domain)
	assert.Equal(t, "aftership", sld)
	assert.Equal(t, "com", tld)
}

// getMD5Hash produces the Gravatar lookup key. Nothing asserted it directly; it
// was only reached through CheckGravatar, which performs an HTTP request. A
// wrong digest would not error anywhere — every Gravatar lookup would silently
// return "no avatar".
//
// The vectors below were computed with an independent md5 implementation rather
// than copied from this code's output, so they actually pin the behaviour.
func TestGetMD5Hash(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  string
	}{
		{"empty", "", "d41d8cd98f00b204e9800998ecf8427e"},
		{"abc", "abc", "900150983cd24fb0d6963f7d28e17f72"},
		{"an address", "someone@example.com", "16d113840f999444259f73bac9ab8b10"},
	}

	for _, c := range cases {
		t.Run(c.name, func(tt *testing.T) {
			got, err := getMD5Hash(c.input)
			require.NoError(tt, err)
			assert.Equal(tt, c.want, got)
		})
	}
}

// The digest is taken over exactly what it is given. Normalising the address is
// CheckGravatar's job, and this pins the split: if the lowering ever moved into
// getMD5Hash, the raw-input vector above would start matching the lowercased one
// and this test would catch it.
func TestGetMD5Hash_DoesNotNormaliseItsInput(t *testing.T) {
	raw := "MyEmailAddress@example.com "
	rawHash, err := getMD5Hash(raw)
	require.NoError(t, err)
	assert.Equal(t, "f9879d71855b5ff21e4963273a886bfc", rawHash)

	normalisedHash, err := getMD5Hash(strings.ToLower(strings.TrimSpace(raw)))
	require.NoError(t, err)
	assert.Equal(t, "0bc83cb571cd1c50ba6f3e8a78ef1346", normalisedHash)

	assert.NotEqual(t, rawHash, normalisedHash,
		"an unnormalised address must not hit the same avatar as a normalised one")
}

func TestGetMD5Hash_Shape(t *testing.T) {
	got, err := getMD5Hash("someone@example.com")
	require.NoError(t, err)
	assert.Len(t, got, 32, "md5 hex is always 32 characters")
	assert.Regexp(t, `^[0-9a-f]{32}$`, got,
		"lowercase hex, since it is concatenated straight onto the avatar URL")
}

func TestGetMD5Hash_IsStableAndDistinct(t *testing.T) {
	first, err := getMD5Hash("alice@example.com")
	require.NoError(t, err)
	second, err := getMD5Hash("alice@example.com")
	require.NoError(t, err)
	assert.Equal(t, first, second, "the same address must map to the same avatar")

	other, err := getMD5Hash("bob@example.com")
	require.NoError(t, err)
	assert.NotEqual(t, first, other, "different addresses must not collide")
}
