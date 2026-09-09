package emailverifier

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

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
		additionalDisposableDomains.Range(func(key, _ interface{}) bool {
			additionalDisposableDomains.Delete(key)
			return true
		})
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

// AddDisposableDomains is reachable from whatever goroutine the caller likes,
// while EnableAutoUpdateDisposable runs updateDisposableDomains on a schedule
// goroutine, and both touch additionalDisposableDomains. While that was a plain
// map the pair was a `fatal error: concurrent map iteration and map write` --
// unrecoverable, not something a caller can defend against.
//
// Nothing caught it because this is the first test in the package to run
// anything concurrently. `make test` does pass -race, but the detector reports
// only conflicting accesses that actually execute, and every other test drives
// both functions from the one goroutine; EnableAutoUpdateDisposable starts a
// schedule, but at a 24h period it never fires inside a test.
func TestAddDisposableDomainsIsConcurrentSafe(t *testing.T) {
	restoreDisposableDomains(t)

	// Two things keep this cheap. updateDisposableDomains walks the whole
	// disposable set, all 75k entries of it, which is beside the point here, so
	// empty it first -- the cleanup above puts it back. And the adder rewrites
	// one key rather than adding new ones: storing an existing key is still a
	// map write, so the race is unchanged, but neither map grows. An adder that
	// keeps adding makes every refresh walk a longer map, and that quadratic
	// cost lands only once the bug is fixed and nothing crashes early.
	disposableSyncDomains.Range(func(key, _ interface{}) bool {
		disposableSyncDomains.Delete(key)
		return true
	})

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, `["fetched.example"]`)
	}))
	defer srv.Close()

	seeded := make([]string, 5000)
	for i := range seeded {
		seeded[i] = fmt.Sprintf("seeded-%d.example", i)
	}
	verifier.AddDisposableDomains(seeded)

	// Both loops run to a deadline rather than a count, so they overlap for the
	// whole window instead of one finishing before the other starts.
	deadline := time.Now().Add(500 * time.Millisecond)
	var wg sync.WaitGroup

	wg.Add(1)
	go func() { // stands in for a caller adding domains as it goes
		defer wg.Done()
		for time.Now().Before(deadline) {
			verifier.AddDisposableDomains([]string{"churned.example"})
		}
	}()

	// Reported after Wait rather than asserted in the goroutine: require calls
	// t.FailNow, which only works on the goroutine running the test. Single
	// writer, and Wait orders the read after it.
	var refreshErr error

	wg.Add(1)
	go func() { // stands in for the schedule goroutine
		defer wg.Done()
		for time.Now().Before(deadline) {
			if refreshErr = updateDisposableDomains(srv.URL); refreshErr != nil {
				return
			}
		}
	}()

	wg.Wait()
	require.NoError(t, refreshErr)

	assert.True(t, verifier.IsDisposable("seeded-0.example"), "a refresh dropped a caller's domain")
	assert.True(t, verifier.IsDisposable("churned.example"))
	assert.True(t, verifier.IsDisposable("fetched.example"))
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
