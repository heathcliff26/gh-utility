package client

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	assert := assert.New(t)

	client := NewClient()
	assert.NotNil(client)
	assert.Equal(defaultEndpoint, client.endpoint)
	assert.NotNil(client.httpClient)
}

func TestGetToken(t *testing.T) {
	t.Run("MissingKey", func(t *testing.T) {
		assert := assert.New(t)

		client := NewClient()
		token, err := client.GetToken("nothing", "abcdf", "12345")
		assert.ErrorContains(err, "failed to read keyfile", "Should return error")
		assert.Empty(token)
	})
	t.Run("MalformedKey", func(t *testing.T) {
		assert := assert.New(t)

		client := NewClient()
		token, err := client.GetToken("testdata/fake-key.txt", "abcdf", "12345")
		assert.ErrorContains(err, "failed to parse keyfile", "Should return error")
		assert.Empty(token)
	})
	t.Run("FailedRequest", func(t *testing.T) {
		assert := assert.New(t)

		client := NewClient()
		client.endpoint = "localhost:6666"

		privateKey := generateRSAKey(t)

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

		client := NewClient()
		client.endpoint = s.URL

		privateKey := generateRSAKey(t)

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

		client := NewClient()
		client.endpoint = s.URL

		privateKey := generateRSAKey(t)

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

		client := NewClient()
		client.endpoint = s.URL

		privateKey := generateRSAKey(t)

		token, err := client.GetToken(privateKey, "abcdf", "12345")
		assert.NoError(err, "Should succeed")
		assert.Equal("abc", token, "Should return token")
	})
}

func generateRSAKey(t *testing.T) string {
	require := require.New(t)

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(err, "Should generate RSA key")

	pemdata := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(key),
		},
	)

	path := filepath.Join(t.TempDir(), "private-key.pem")
	err = os.WriteFile(path, pemdata, 0600)
	require.NoError(err, "Should write RSA key")

	return path
}
