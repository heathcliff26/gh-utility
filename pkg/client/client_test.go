package client

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/heathcliff26/gh-utility/testutils"
	"github.com/stretchr/testify/assert"
)

func TestNewClient(t *testing.T) {
	assert := assert.New(t)

	client := NewClient(DefaultEndpoint)
	assert.NotNil(client)
	assert.Equal(DefaultEndpoint, client.endpoint)
	assert.NotNil(client.httpClient)
	assert.Equal(defaultClientTimeout, client.httpClient.Timeout)
}

func TestGetToken(t *testing.T) {
	t.Run("MissingKey", func(t *testing.T) {
		assert := assert.New(t)

		client := NewClient("http://localhost")
		token, err := client.GetToken("nothing", "abcdf", "12345")
		assert.ErrorContains(err, "failed to read keyfile", "Should return error")
		assert.Empty(token)
	})
	t.Run("MalformedKey", func(t *testing.T) {
		assert := assert.New(t)

		client := NewClient("http://localhost")
		token, err := client.GetToken("testdata/fake-key.txt", "abcdf", "12345")
		assert.ErrorContains(err, "failed to parse keyfile", "Should return error")
		assert.Empty(token)
	})
	t.Run("FailedRequest", func(t *testing.T) {
		assert := assert.New(t)

		client := NewClient("http://localhost:6666")

		privateKey := testutils.GenerateRSAKey(t)

		token, err := client.GetToken(privateKey, "abcdf", "12345")
		assert.ErrorContains(err, "failed to send request", "Should return error")
		assert.Empty(token)
	})
	t.Run("RequestReturnsFailure", func(t *testing.T) {
		assert := assert.New(t)

		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer s.Close()

		client := NewClient(s.URL)

		privateKey := testutils.GenerateRSAKey(t)

		token, err := client.GetToken(privateKey, "abcdf", "12345")
		assert.ErrorContains(err, "request returned non-201 status", "Should return error")
		assert.Empty(token)
	})
	t.Run("ParseResponseFailure", func(t *testing.T) {
		assert := assert.New(t)

		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte("not json"))
		}))
		defer s.Close()

		client := NewClient(s.URL)

		privateKey := testutils.GenerateRSAKey(t)

		token, err := client.GetToken(privateKey, "abcdf", "12345")
		assert.ErrorContains(err, "failed to decode response", "Should return error")
		assert.Empty(token)
	})
	t.Run("Success", func(t *testing.T) {
		assert := assert.New(t)

		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(http.MethodPost, r.Method, "Should use POST method")
			assert.Equal("/app/installations/12345/access_tokens", r.URL.Path, "Should use correct path")
			assert.NotEmpty(r.Header.Get("Accept"), "Should set Accept header")

			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"token": "abc", "expires_at": "future"}`))
		}))
		defer s.Close()

		client := NewClient(s.URL)

		privateKey := testutils.GenerateRSAKey(t)

		token, err := client.GetToken(privateKey, "abcdf", "12345")
		assert.NoError(err, "Should succeed")
		assert.Equal("abc", token, "Should return token")
	})
}
