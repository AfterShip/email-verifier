package emailverifier

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"
)

// restoreDisposableDomains puts the package-level set back once the test ends.
// updateDisposableDomains replaces it wholesale, and nothing used to undo
// that: every test running after one of these saw the mock list instead of the
// generated data, which is how two fixtures came to assert against domains
// that were never in the generated maps at all.
func restoreDisposableDomains(t *testing.T) {
	t.Cleanup(func() {
		disposableSyncDomains.Range(func(key, _ interface{}) bool {
			disposableSyncDomains.Delete(key)
			return true
		})
		for d := range disposableDomains {
			disposableSyncDomains.Store(d, struct{}{})
		}
		for d := range additionalDisposableDomains {
			delete(additionalDisposableDomains, d)
		}
	})
}

func TestUpdateDisposableDomainsOK(t *testing.T) {
	restoreDisposableDomains(t)

	assert.False(t, verifier.IsDisposable("a.org"))
	assert.False(t, verifier.IsDisposable("b.com"))

	assert.True(t, verifier.IsDisposable("mailinator.com"))

	mockResp := []string{"a.org", "b.com", "zzjbfwqi.shop", "dbbd8.club"}
	defer gock.Off()
	gock.New("https://raw.githubusercontent.com").
		Get("/disposable/disposable-email-domains/master/domains.json").
		Reply(http.StatusOK).
		JSON(mockResp)

	err := updateDisposableDomains(disposableDataURL)
	assert.NoError(t, err)
	assert.True(t, verifier.IsDisposable("a.org"))
	assert.True(t, verifier.IsDisposable("b.com"))
	assert.False(t, verifier.IsDisposable("c.net"))
	assert.False(t, verifier.IsDisposable("mailinator.com"))
}

// The upstream list has these domains; disposable_allowlist.txt is the only
// thing keeping them out, so a refresh that reinstated them would quietly undo
// what the generated list was built to express.
func TestUpdateDisposableDomainsKeepsAllowlistOut(t *testing.T) {
	restoreDisposableDomains(t)

	mockResp := []string{"a.org", "hush.com", "lavabit.com"}
	defer gock.Off()
	gock.New("https://raw.githubusercontent.com").
		Get("/disposable/disposable-email-domains/master/domains.json").
		Reply(http.StatusOK).
		JSON(mockResp)

	require.NoError(t, updateDisposableDomains(disposableDataURL))

	assert.True(t, verifier.IsDisposable("a.org"))
	assert.False(t, verifier.IsDisposable("hush.com"))
	assert.False(t, verifier.IsDisposable("lavabit.com"))
}

// A caller who deliberately blocks one of the allowlisted domains outranks the
// allowlist, which is only there to correct the upstream list.
func TestAddDisposableDomainsOutranksAllowlist(t *testing.T) {
	restoreDisposableDomains(t)

	verifier.AddDisposableDomains([]string{"hush.com"})

	mockResp := []string{"a.org", "hush.com"}
	defer gock.Off()
	gock.New("https://raw.githubusercontent.com").
		Get("/disposable/disposable-email-domains/master/domains.json").
		Reply(http.StatusOK).
		JSON(mockResp)

	require.NoError(t, updateDisposableDomains(disposableDataURL))

	assert.True(t, verifier.IsDisposable("hush.com"))
}

func TestUpdateDisposableDomainsFailed_NoSuchHost(t *testing.T) {

	err := updateDisposableDomains("http://abcmockxyz.aaa")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no such host")
}

func TestUpdateDisposableDomainsFailed_StatusNotFound(t *testing.T) {
	defer gock.Off()
	gock.New("https://raw.githubusercontent.com").
		Get("/disposable/disposable-email-domains/master/domains.json").
		Reply(http.StatusNotFound)

	err := updateDisposableDomains(disposableDataURL)
	assert.Error(t, err, "get disposable domains from https://raw.githubusercontent.com/disposable/disposable-email-domains/master/domains.json with status_code: 404")
}

func TestUpdateDisposableDomainsFailed_StatusInternalError(t *testing.T) {
	defer gock.Off()
	gock.New("https://raw.githubusercontent.com").
		Get("/disposable/disposable-email-domains/master/domains.json").
		Reply(http.StatusInternalServerError)

	err := updateDisposableDomains(disposableDataURL)
	assert.Error(t, err, "get disposable domains from https://raw.githubusercontent.com/disposable/disposable-email-domains/master/domains.json with status_code: 500")
}

func TestUpdateDisposableDomains_NoResponse(t *testing.T) {

	defer gock.Off()
	gock.New("https://raw.githubusercontent.com").
		Get("/disposable/disposable-email-domains/master/domains.json").
		Reply(http.StatusOK)

	err := updateDisposableDomains(disposableDataURL)
	assert.NoError(t, err)
}

func TestUpdateDisposableDomains_WrongResponse(t *testing.T) {

	defer gock.Off()
	gock.New("https://raw.githubusercontent.com").
		Get("/disposable/disposable-email-domains/master/domains.json").
		Reply(http.StatusOK).
		JSON("testing")

	err := updateDisposableDomains(disposableDataURL)
	assert.Error(t, err, "invalid character 'e' in literal true (expecting 'r')")
}
