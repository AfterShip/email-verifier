package emailverifier

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// GenerateRandomEmail feeds the catch-all probe: every SMTP exchange sends RCPT
// TO for this address and reads the reply as "does this domain accept anything".
// Nothing asserted its output before, so its shape was only exercised
// incidentally through CheckSMTP, which needs a live mail server.
func TestGenerateRandomEmail(t *testing.T) {
	const domain = "example.com"

	got := GenerateRandomEmail(domain)

	local, host, found := strings.Cut(got, "@")
	require.True(t, found, "address must contain an @: %q", got)
	assert.Equal(t, domain, host, "the domain is passed through unchanged")
	assert.Len(t, local, 32, "the local part is a fixed 32 characters")

	// The probe is only meaningful if the address cannot collide with a real
	// mailbox. A short or guessable local part risks matching one, which would
	// read as "this domain is catch-all" and mark every address on it
	// unverifiable.
	assert.Regexp(t, `^[a-z0-9]+$`, local, "local part must stay inside alphanumeric")
}

func TestGenerateRandomEmail_IsNotDeterministic(t *testing.T) {
	// A fixed sequence would let a provider recognise the probe and answer
	// differently for it than for a real recipient.
	seen := make(map[string]struct{}, 100)
	for range 100 {
		seen[GenerateRandomEmail("example.com")] = struct{}{}
	}
	assert.Greater(t, len(seen), 90, "100 draws should almost all be distinct")
}

func TestGenerateRandomEmail_KeepsTheDomainVerbatim(t *testing.T) {
	// Subdomains and mixed case must survive, because the address is handed
	// straight to RCPT TO and a mangled domain would be rejected for the wrong
	// reason.
	for _, domain := range []string{"mail.example.co.uk", "Example.COM", "xn--fsq.example"} {
		got := GenerateRandomEmail(domain)
		assert.True(t, strings.HasSuffix(got, "@"+domain), "%q should end with @%s", got, domain)
	}
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
