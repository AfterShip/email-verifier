package emailverifier

import (
	"os"
	"testing"
)

// requireLiveNetwork keeps unit-test runs independent from third-party services
// and from networks that block outbound SMTP. Set EMAIL_VERIFIER_INTEGRATION=1
// to run the live integration checks explicitly.
func requireLiveNetwork(t *testing.T) {
	t.Helper()
	if os.Getenv("EMAIL_VERIFIER_INTEGRATION") != "1" {
		t.Skip("live network test; set EMAIL_VERIFIER_INTEGRATION=1 to run")
	}
}
