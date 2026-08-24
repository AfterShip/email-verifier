package emailverifier

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestYahooCheckByAPI(t *testing.T) {
	client := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		resp := &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Request:    req,
		}
		switch req.Method {
		case http.MethodGet:
			resp.Header.Add("Set-Cookie", "AS=v=1&s=test-acrumb&d=value; Path=/")
			resp.Body = io.NopCloser(strings.NewReader(`<input value="test-session" name="sessionIndex">`))
		case http.MethodPost:
			body, err := io.ReadAll(req.Body)
			if err != nil {
				return nil, err
			}
			if strings.Contains(string(body), `"userId":"hello"`) {
				resp.Body = io.NopCloser(strings.NewReader(`{"errors":[{"name":"userId","error":"IDENTIFIER_EXISTS"}]}`))
			} else {
				resp.Body = io.NopCloser(strings.NewReader(`{"errors":[]}`))
			}
		}
		return resp, nil
	})}
	yahooAPIVerifier := newYahooAPIVerifier(client)
	t.Run("email exists", func(tt *testing.T) {
		res, err := yahooAPIVerifier.check("yahoo.com", "hello")
		require.NoError(tt, err)
		assert.True(tt, res.HostExists)
		assert.True(tt, res.Deliverable)
	})
	t.Run("invalid email not exists", func(tt *testing.T) {
		res, err := yahooAPIVerifier.check("yahoo.com", "123")
		require.NoError(tt, err)
		assert.True(tt, res.HostExists)
		assert.False(tt, res.Deliverable)
	})
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
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
