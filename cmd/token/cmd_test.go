package token

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/heathcliff26/gh-utility/testutils"
	"github.com/stretchr/testify/assert"
)

func TestRequiredFlags(t *testing.T) {
	assert := assert.New(t)

	cmd := NewCommand()

	err := cmd.ValidateRequiredFlags()
	assert.ErrorContains(err, "required flag(s)")
	assert.ErrorContains(err, appKeyPathFlag)
	assert.ErrorContains(err, clientIDFlag)
	assert.ErrorContains(err, installationIDFlag)
}

func TestRun(t *testing.T) {
	t.Run("Error", func(t *testing.T) {
		assert := assert.New(t)

		cmd := NewCommand()

		err := run(cmd)
		assert.Error(err, "Should fail")
	})
	t.Run("ConsoleOutput", func(t *testing.T) {
		assert := assert.New(t)

		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(http.MethodPost, r.Method, "Should use POST method")
			assert.Equal("/app/installations/12345/access_tokens", r.URL.Path, "Should use correct path")
			assert.NotEmpty(r.Header.Get("Accept"), "Should set Accept header")

			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"token": "abc", "expires_at": "future"}`))
		}))
		defer s.Close()

		privateKey := testutils.GenerateRSAKey(t)

		args := []string{
			"--" + appKeyPathFlag, privateKey,
			"--" + clientIDFlag, "abcdf",
			"--" + installationIDFlag, "12345",
			"--" + endpointFlag, s.URL,
		}
		cmd := NewCommand()
		_ = cmd.ParseFlags(args)

		var buf bytes.Buffer
		w := io.Writer(&buf)
		cmd.SetOut(w)

		err := run(cmd)
		assert.NoError(err, "Should succeed")
		assert.Equal("abc\n", buf.String(), "Should output token to console")
	})
	t.Run("FileOutput", func(t *testing.T) {
		assert := assert.New(t)

		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(http.MethodPost, r.Method, "Should use POST method")
			assert.Equal("/app/installations/12345/access_tokens", r.URL.Path, "Should use correct path")
			assert.NotEmpty(r.Header.Get("Accept"), "Should set Accept header")

			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"token": "abc", "expires_at": "future"}`))
		}))
		defer s.Close()

		privateKey := testutils.GenerateRSAKey(t)

		outputFile := filepath.Join(t.TempDir(), "token.txt")

		args := []string{
			"--" + appKeyPathFlag, privateKey,
			"--" + clientIDFlag, "abcdf",
			"--" + installationIDFlag, "12345",
			"--" + endpointFlag, s.URL,
			"--" + outputFlag, outputFile,
		}
		cmd := NewCommand()
		_ = cmd.ParseFlags(args)

		var buf bytes.Buffer
		w := io.Writer(&buf)
		cmd.SetOut(w)

		err := run(cmd)
		assert.NoError(err, "Should succeed")
		assert.Empty(buf.String(), "Should not output token to console")
		token, err := os.ReadFile(outputFile)
		assert.NoError(err, "Should be able to read output file")
		assert.Equal("abc", string(token), "Should output token to file")
	})
}
