package emailverifier

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestYahooCheckByAPI(t *testing.T) {
	yahooAPIVerifier := newYahooAPIVerifier(nil)
	t.Run("email exists", func(tt *testing.T) {
		res, err := yahooAPIVerifier.check("yahoo.com", "hello")
		require.NoError(tt, err)
		assert.Equal(tt, true, res.HostExists)
		assert.Equal(tt, true, res.Deliverable)
	})
	t.Run("invalid email not exists", func(tt *testing.T) {
		res, err := yahooAPIVerifier.check("yahoo.com", "123")
		require.NoError(tt, err)
		assert.Equal(tt, true, res.HostExists)
		assert.Equal(tt, false, res.Deliverable)
	})
}

func TestGetAcrumb(t *testing.T) {
	cookies0 := []*http.Cookie{
		{Value: "123321"},
		{Value: "v=1&s=gWKqrs5c&d=A6454c24b|Zt.ZFgb.2T"},
	}
	acrumb := getAcrumb(cookies0)
	assert.Equal(t, "gWKqrs5c", acrumb)

	cookies1 := []*http.Cookie{
		{Value: "123321"},
		{Value: "v=1&s=gWKqrs5c"},
	}
	acrumb = getAcrumb(cookies1)
	assert.Equal(t, "gWKqrs5c", acrumb)

	cookies2 := []*http.Cookie{
		{Value: "123321"},
	}
	acrumb = getAcrumb(cookies2)
	assert.Empty(t, acrumb)
}
