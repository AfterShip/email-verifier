package emailverifier

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/h2non/gock.v1"
)

func TestUpdateDisposableDomainsOK(t *testing.T) {
	// updateDisposableDomains replaces the package-level set, and nothing here
	// used to put it back: every test that ran afterwards saw only the four
	// domains mocked below, so TestIsDisposableDomain_True was passing because
	// its fixture appeared in that mock rather than in the generated data.
	t.Cleanup(func() {
		disposableSyncDomains.Range(func(key, _ interface{}) bool {
			disposableSyncDomains.Delete(key)
			return true
		})
		for d := range disposableDomains {
			disposableSyncDomains.Store(d, struct{}{})
		}
	})

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
