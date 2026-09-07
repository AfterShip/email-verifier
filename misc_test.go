package emailverifier

import (
	"bufio"
	"os"
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
