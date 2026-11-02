package emailverifier_test

import (
	"testing"

	emailverifier "github.com/AfterShip/email-verifier"
)

// Verifier is exported and its setters are exported, so a composite literal is
// legal from outside the package even though NewVerifier is the intended
// constructor. It reached the network before the newClient seam existed, and
// must not start panicking because of it.
func TestZeroVerifierDoesNotPanic(t *testing.T) {
	v := (&emailverifier.Verifier{}).
		EnableSMTPCheck().
		DisableCatchAllCheck().
		ConnectTimeout(1)

	_, err := v.CheckSMTP("domain-that-does-not-resolve.invalid", "user")
	if err == nil {
		t.Fatal("expected the lookup to fail, not to succeed")
	}
}
