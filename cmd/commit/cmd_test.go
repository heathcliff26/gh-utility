package commit

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/heathcliff26/gh-utility/testutils"
	"github.com/stretchr/testify/assert"
)

func TestRun(t *testing.T) {
	t.Run("NoChanges", func(t *testing.T) {
		assert := assert.New(t)

		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Log("Should not make any api calls")
			t.Fail()

			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer s.Close()

		dir, _ := testutils.NewTestRepository(t)

		cmd := NewCommand()

		var buf bytes.Buffer
		w := io.Writer(&buf)
		cmd.SetOut(w)

		err := run(cmd, dir, "testbranch", "Test Commit Message", s.URL)
		assert.NoError(err, "Should succeed")
		assert.Equal("No changes detected, exiting\n", buf.String(), "Should exit early")
	})
	t.Run("NoRepository", func(t *testing.T) {
		assert := assert.New(t)

		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Log("Should not make any api calls")
			t.Fail()

			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer s.Close()

		cmd := NewCommand()

		err := run(cmd, t.TempDir(), "testbranch", "Test Commit Message", s.URL)
		assert.Error(err, "Should not succeed")
	})
}
