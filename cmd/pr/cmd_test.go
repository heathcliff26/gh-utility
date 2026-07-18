package pr

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

		args := []string{
			"--" + repositoryFlag, dir,
			"--" + branchFlag, "testbranch",
			"--" + titleFlag, "Test PR",
			"--" + bodyFlag, "Lorem Ipsum",
			"--" + endpointFlag, s.URL,
		}
		cmd := NewCommand()
		_ = cmd.ParseFlags(args)

		var buf bytes.Buffer
		w := io.Writer(&buf)
		cmd.SetOut(w)

		err := run(cmd)
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

		args := []string{
			"--" + repositoryFlag, t.TempDir(),
			"--" + branchFlag, "testbranch",
			"--" + titleFlag, "Test PR",
			"--" + bodyFlag, "Lorem Ipsum",
			"--" + endpointFlag, s.URL,
		}
		cmd := NewCommand()
		_ = cmd.ParseFlags(args)

		err := run(cmd)
		assert.Error(err, "Should not succeed")
	})
}
