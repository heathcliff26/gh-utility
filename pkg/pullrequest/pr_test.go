package pullrequest

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-git/go-git/v5/config"
	"github.com/heathcliff26/gh-utility/pkg/client"
	"github.com/heathcliff26/gh-utility/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCommit(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		assert := assert.New(t)
		require := require.New(t)

		dir, repo := testutils.NewTestRepository(t)
		_, err := repo.CreateRemote(&config.RemoteConfig{
			Name: "origin",
			URLs: []string{"https://github.com/heathcliff26/gh-utility.git"},
		})
		require.NoError(err, "Should add remote")

		err = os.WriteFile(filepath.Join(dir, "new.txt"), []byte("new content"), 0644)
		require.NoError(err, "Should create change")

		var requestCount int
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch requestCount {
			case 0:
				assert.Equal(http.MethodPost, r.Method, "Should create tree")
				assert.Equal("/repos/heathcliff26/gh-utility/git/trees", r.URL.Path, "Should use tree endpoint")
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
			default:
				t.Fatalf("unexpected request %d", requestCount)
			}
			requestCount++
		}))
		defer s.Close()

		c := client.NewClient(s.URL)

		treeHash, commitHash, err := Commit(c, dir, "test commit", "feature", "token")
		require.NoError(err, "Should commit changes")
		assert.Equal("tree-sha", treeHash, "Should return tree hash")
		assert.Equal("commit-sha", commitHash, "Should return commit hash")
		assert.Equal(4, requestCount, "Should make expected API calls")
	})
}

func TestPullRequest(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		assert := assert.New(t)
		require := require.New(t)

		dir, repo := testutils.NewTestRepository(t)
		_, err := repo.CreateRemote(&config.RemoteConfig{
			Name: "origin",
			URLs: []string{"https://github.com/heathcliff26/gh-utility.git"},
		})
		require.NoError(err, "Should add remote")

		headRef, err := repo.Head()
		require.NoError(err, "Should read current branch")
		baseBranch := strings.TrimPrefix(headRef.Name().String(), "refs/heads/")

		var requestCount int
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch requestCount {
			case 0:
				assert.Equal(http.MethodGet, r.Method, "Should list pull requests")
				assert.Equal("/repos/heathcliff26/gh-utility/pulls", r.URL.Path, "Should use pulls endpoint")
				assert.Equal("heathcliff26:feature", r.URL.Query().Get("head"), "Should query for the head branch")
				_, _ = w.Write([]byte(`[]`))
			case 1:
				assert.Equal(http.MethodPost, r.Method, "Should create pull request")
				assert.Equal("/repos/heathcliff26/gh-utility/pulls", r.URL.Path, "Should use pulls endpoint")

				var prReq client.PrRequest
				require.NoError(json.NewDecoder(r.Body).Decode(&prReq), "Should decode pull request payload")
				assert.Equal("feature title", prReq.Title, "Should use supplied title")
				assert.Equal("feature body", prReq.Body, "Should use supplied body")
				assert.Equal("heathcliff26:feature", prReq.Head, "Should use supplied head branch")
				assert.Equal(baseBranch, prReq.Base, "Should use current branch as base")

				w.WriteHeader(http.StatusCreated)
				_, _ = w.Write([]byte(`{"html_url":"https://example.com/pr/1"}`))
			default:
				t.Fatalf("unexpected request %d", requestCount)
			}
			requestCount++
		}))
		defer s.Close()

		c := client.NewClient(s.URL)

		prURL, err := PullRequest(c, dir, "feature", "feature title", "feature body", "token", nil)
		require.NoError(err, "Should create pull request")
		assert.Equal("https://example.com/pr/1", prURL, "Should return pull request URL")
		assert.Equal(2, requestCount, "Should make expected API calls")
	})
}
