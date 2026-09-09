package emailverifier

import (
	"bufio"
	"os"
	"regexp"
	"strconv"
	"strings"
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
		{"disposable allowlist", "cmd/build_metadata/disposable_allowlist.txt", allowedDisposableDomains},
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
				entry := strings.TrimSpace(scanner.Text())
				// Skipped the same way the generator skips them, so a
				// hand-maintained list can carry its own reasoning.
				if entry == "" || strings.HasPrefix(entry, "#") {
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
// EnableAutoUpdateDisposable refetches the list at disposableDataURL and, via
// updateDisposableDomains, deletes every baked-in domain the fetched list
// omits. So the list cmd/build_metadata bakes in has to be the list this
// constant points at. It was not: the script built from tompec while this
// pointed at disposable/, lists sharing only 37585 of tompec's 133602
// entries, so turning auto-update on dropped 96017 domains and added 37679.
func TestUpdateScriptFetchesTheDisposableDataURL(t *testing.T) {
	script, err := os.ReadFile("cmd/build_metadata/update.sh")
	require.NoError(t, err)

	assert.Contains(t, string(script), disposableDataURL,
		"cmd/build_metadata/update.sh must bake in the same disposable list that disposableDataURL refreshes at runtime")
}

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
// dead weight that no test notices, and both maps shipped some. free.txt
// carried a U+00A0 no-break space on the end of atlanticbb.net for years:
// one upstream entry has one attached, and LC_ALL=C [[:space:]] does not
// match it, so every cleanup in update.sh walked past it and
// IsFreeDomain("atlanticbb.net") returned false. The disposable list carried
// twelve IDNs spelled in Unicode rather than punycode, dead for a related
// reason -- IsDisposable converts its argument with domainToASCII before the
// lookup, so nothing it was given could ever match them.
func TestGeneratedDomainsAreWellFormedHostnames(t *testing.T) {
	wellFormed := regexp.MustCompile(`^[a-z0-9]([a-z0-9-]*[a-z0-9])?(\.[a-z0-9]([a-z0-9-]*[a-z0-9])?)+$`)

	for name, generated := range map[string]map[string]bool{
		"free":       freeDomains,
		"disposable": disposableDomains,
	} {
		var malformed []string
		for domain := range generated {
			if !wellFormed.MatchString(domain) {
				if len(malformed) < 10 {
					malformed = append(malformed, strconv.Quote(domain))
				}
			}
		}

		assert.Emptyf(t, malformed, "%s domains that no lookup can match; check the normalisation in cmd/build_metadata/update.sh", name)
	}
}

// disposable_allowlist.txt exists because the upstream disposable list
// classifies a handful of real providers as throwaway services. update.sh
// subtracts it, so nothing in it should reach the generated map -- and the
// subtraction matters twice, since free.txt is the free candidates minus the
// disposable list: a domain wrongly listed there is not merely reported
// disposable, it is also dropped from free, leaving the library answering
// "neither" about a real mailbox provider.
func TestAllowedDomainsAreNotDisposable(t *testing.T) {
	require.NotEmpty(t, allowedDisposableDomains, "allowlist is empty; the generated map is stale")

	var listed []string
	for domain := range allowedDisposableDomains {
		if disposableDomains[domain] {
			listed = append(listed, domain)
		}
	}

	assert.Empty(t, listed, "allowlisted domains still in the generated disposable map; check the subtraction in cmd/build_metadata/update.sh")
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
	// A long-lived service rather than one of the churning throwaway domains:
	// the previous fixture, dbbd8.club, was only ever in the list this repo
	// used to build from and vanished when the two were reconciled.
	domain := "mailinator.com"

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
