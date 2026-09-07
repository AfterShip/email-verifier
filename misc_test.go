package emailverifier

import (
	"bufio"
	"os"
	"regexp"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var verifier = NewVerifier().EnableSMTPCheck()

// The metadata_*.go files are generated from the .txt lists under
// cmd/build_metadata, and nothing kept the two in step. They drifted: cto,
// ctos, cfo and cfos were added to role.txt in December 2025 and
// metadata_role.go was never regenerated, so IsRoleAccount("cto") returned
// false for the nine months that followed. This fails if a list is edited
// without regenerating.
func TestGeneratedMetadataMatchesSources(t *testing.T) {
	cases := []struct {
		name      string
		source    string
		generated map[string]bool
	}{
		{"disposable", "cmd/build_metadata/disposable.txt", disposableDomains},
		{"free", "cmd/build_metadata/free.txt", freeDomains},
		{"role", "cmd/build_metadata/role.txt", roleAccounts},
	}

	for _, c := range cases {
		test := c
		t.Run(test.name, func(tt *testing.T) {
			file, err := os.Open(test.source)
			require.NoError(tt, err)
			defer func() { assert.NoError(tt, file.Close()) }()

			// The generator skips duplicates, so compare against the set of
			// distinct entries rather than the line count.
			entries := make(map[string]bool)
			var missing []string
			scanner := bufio.NewScanner(file)
			scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
			for scanner.Scan() {
				entry := scanner.Text()
				if entry == "" {
					continue
				}
				entries[entry] = true
				if !test.generated[entry] && len(missing) < 10 {
					missing = append(missing, entry)
				}
			}
			require.NoError(tt, scanner.Err())
			require.NotEmpty(tt, entries, "source list is empty")

			assert.Emptyf(tt, missing, "%s has entries missing from the generated map; run `cd cmd/build_metadata && go run main.go`", test.source)
			// Compare counts rather than the maps themselves; asserting on the
			// containers dumps six figures of entries into the failure output.
			sourceCount, generatedCount := len(entries), len(test.generated)
			assert.Equalf(tt, sourceCount, generatedCount, "%s holds %d distinct entries but the generated map holds %d", test.source, sourceCount, generatedCount)
		})
	}
}

// Every free source is a signup blocklist rather than a directory of
// providers -- HubSpot publishes theirs under Marketing/Lead-Capture -- so
// they carry throwaway domains next to the real ones and update.sh subtracts
// the disposable list from them. That subtraction is the only thing keeping
// the two apart, and getting it wrong is not loud: consuming a free list that
// had merged two disposable blocklists into itself put ~5k throwaway domains,
// whole families such as aws-mail-free-<n>.dynv6.net, into free.txt, where
// IsFreeDomain called them free and SuggestDomain offered them as
// corrections. Assert the property the subtraction provides.
func TestFreeAndDisposableDomainsAreDisjoint(t *testing.T) {
	var overlap []string
	for domain := range freeDomains {
		if disposableDomains[domain] {
			if len(overlap) < 10 {
				overlap = append(overlap, domain)
			}
		}
	}

	assert.Empty(t, overlap, "domains classified as both free and disposable; check the subtraction in cmd/build_metadata/update.sh")
}

// A key that is not a hostname can never be returned by a lookup, so it is
// dead weight that no test notices. free.txt shipped "atlanticbb.net "
// for years -- one upstream entry has a no-break space attached, and
// LC_ALL=C [[:space:]] does not match U+00A0, so every cleanup in update.sh
// walked past it and IsFreeDomain("atlanticbb.net") returned false.
//
// disposableDomains is not checked here: twelve of its entries are IDNs
// spelled in Unicode instead of punycode, and since IsDisposable converts its
// argument with domainToASCII first, those keys are dead in the same way.
// Converting them is a separate change to the disposable pipeline.
func TestFreeDomainsAreWellFormedHostnames(t *testing.T) {
	wellFormed := regexp.MustCompile(`^[a-z0-9]([a-z0-9-]*[a-z0-9])?(\.[a-z0-9]([a-z0-9-]*[a-z0-9])?)+$`)

	var malformed []string
	for domain := range freeDomains {
		if !wellFormed.MatchString(domain) {
			if len(malformed) < 10 {
				malformed = append(malformed, strconv.Quote(domain))
			}
		}
	}

	assert.Empty(t, malformed, "free domains that no lookup can match; check the normalisation in cmd/build_metadata/update.sh")
}

func TestIsFreeDomain_True(t *testing.T) {
	domain := "gmail.com"

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
