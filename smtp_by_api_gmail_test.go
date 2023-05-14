package emailverifier

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGmailCheckByAPI(t *testing.T) {
	gmailAPIVerifier := newGmailAPIVerifier(nil)

	t.Run("email exists", func(tt *testing.T) {
		res, err := gmailAPIVerifier.check("gmail.com", "someone")
		assert.NoError(t, err)
		assert.Equal(t, true, res.HostExists)
		assert.Equal(t, true, res.Deliverable)
	})
	t.Run("invalid email not exists", func(tt *testing.T) {
		// username must greater than 6 characters
		res, err := gmailAPIVerifier.check("gmail.com", "hello")
		assert.NoError(t, err)
		assert.Equal(t, true, res.HostExists)
		assert.Equal(t, false, res.Deliverable)
	})
}
