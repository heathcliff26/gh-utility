package pr

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-git/go-git/v5/config"
	"github.com/heathcliff26/gh-utility/pkg/client"
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
	t.Run("WithLabels", func(t *testing.T) {
		assert := assert.New(t)

		dir, repo := testutils.NewTestRepository(t)
		_, err := repo.CreateRemote(&config.RemoteConfig{
			Name: "origin",
			URLs: []string{"https://github.com/heathcliff26/gh-utility.git"},
		})
		assert.NoError(err, "Should add remote")

		err = os.WriteFile(filepath.Join(dir, "new.txt"), []byte("new content"), 0644)
		assert.NoError(err, "Should create change")

		var requestCount int
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch requestCount {
			case 0:
				assert.Equal(http.MethodPost, r.Method, "Should create tree")
				assert.Equal("/repos/heathcliff26/gh-utility/git/trees", r.URL.Path, "Should use tree endpoint")

				var req client.TreeRequest
				err := json.NewDecoder(r.Body).Decode(&req)
				assert.NoError(err, "Should decode tree request")
				assert.Len(req.Tree, 1, "Should have 1 tree object")

				w.WriteHeader(http.StatusCreated)
				_, _ = w.Write([]byte(`{"sha":"tree-sha"}`))
			case 1:
				assert.Equal(http.MethodPost, r.Method, "Should create commit")
				assert.Equal("/repos/heathcliff26/gh-utility/git/commits", r.URL.Path, "Should use commit endpoint")
				w.WriteHeader(http.StatusCreated)
				_, _ = w.Write([]byte(`{"sha":"commit-sha"}`))
			case 2:
				assert.Equal(http.MethodGet, r.Method, "Should check branch existence")
				assert.Equal("/repos/heathcliff26/gh-utility/git/refs/heads/feature", r.URL.Path, "Should use branch ref endpoint")
				w.WriteHeader(http.StatusNotFound)
			case 3:
				assert.Equal(http.MethodPost, r.Method, "Should create branch")
				assert.Equal("/repos/heathcliff26/gh-utility/git/refs", r.URL.Path, "Should use refs endpoint")
				w.WriteHeader(http.StatusCreated)
			case 4:
				assert.Equal(http.MethodGet, r.Method, "Should list pull requests")
				assert.Equal("/repos/heathcliff26/gh-utility/pulls", r.URL.Path, "Should use pulls endpoint")
				_, _ = w.Write([]byte(`[]`))
			case 5:
				assert.Equal(http.MethodPost, r.Method, "Should create pull request")
				assert.Equal("/repos/heathcliff26/gh-utility/pulls", r.URL.Path, "Should use pulls endpoint")
				w.WriteHeader(http.StatusCreated)
				_, _ = w.Write([]byte(`{"html_url":"https://example.com/pr/1","number":1,"labels":[]}`))
			case 6:
				assert.Equal(http.MethodPost, r.Method, "Should add labels")
				assert.Equal("/repos/heathcliff26/gh-utility/issues/1/labels", r.URL.Path, "Should use issues labels endpoint")

				var labelReq client.LabelRequest
				err := json.NewDecoder(r.Body).Decode(&labelReq)
				assert.NoError(err, "Should decode label request")
				assert.Equal([]string{"bug", "enhancement"}, labelReq.Labels, "Should have correct labels")

				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`[{"name":"bug"},{"name":"enhancement"}]`))
			default:
				t.Fatalf("unexpected request %d", requestCount)
			}
			requestCount++
		}))
		defer s.Close()

		args := []string{
			"--" + repositoryFlag, dir,
			"--" + branchFlag, "feature",
			"--" + titleFlag, "Test PR",
			"--" + bodyFlag, "Lorem Ipsum",
			"--" + labelFlag, "bug",
			"--" + labelFlag, "enhancement",
			"--" + endpointFlag, s.URL,
		}
		cmd := NewCommand()
		_ = cmd.ParseFlags(args)

		var buf bytes.Buffer
		w := io.Writer(&buf)
		cmd.SetOut(w)

		err = run(cmd)
		assert.NoError(err, "Should succeed")
		assert.Equal("Tree: tree-sha\nCommit: commit-sha\n\nhttps://example.com/pr/1\n", buf.String(), "Should print PR URL")
		assert.Equal(7, requestCount, "Should make expected API calls")
	})
}
