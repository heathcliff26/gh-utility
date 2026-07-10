package git

import (
	"os"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCloneOptions(t *testing.T) {
	tMatrix := []struct {
		Name       string
		RepoURL    string
		FetchDepth int
		Token      string
		Result     *git.CloneOptions
	}{
		{
			Name:       "Basic",
			RepoURL:    "https://github.com/foo/bar.git",
			FetchDepth: 0,
			Token:      "",
			Result: &git.CloneOptions{
				URL:   "https://github.com/foo/bar.git",
				Depth: 0,
			},
		},
		{
			Name:       "WithToken",
			RepoURL:    "https://github.com/foo/bar.git",
			FetchDepth: 0,
			Token:      "testtoken",
			Result: &git.CloneOptions{
				URL:   "https://github.com/foo/bar.git",
				Depth: 0,
				Auth: &http.BasicAuth{
					Username: "x-access-token",
					Password: "testtoken",
				},
			},
		},
	}

	for _, tCase := range tMatrix {
		t.Run(tCase.Name, func(t *testing.T) {
			assert := assert.New(t)

			if tCase.Token != "" {
				t.Setenv("GITHUB_TOKEN", tCase.Token)
			}

			result := newCloneOptions(tCase.RepoURL, tCase.FetchDepth)

			assert.Equal(tCase.Result, result, "Should return expected result")
		})
	}
}

func TestExtractRepositoryName(t *testing.T) {
	tMatrix := []struct {
		Name   string
		URL    string
		Result string
	}{
		{
			Name:   "HTTP",
			URL:    "https://github.com/heathcliff26/gh-utility.git",
			Result: "gh-utility",
		},
		{
			Name:   "SSH",
			URL:    "git@github.com:heathcliff26/gh-utility.git",
			Result: "gh-utility",
		},
		{
			Name:   "NoGitSuffix",
			URL:    "https://github.com/heathcliff26/gh-utility",
			Result: "gh-utility",
		},
		{
			Name:   "EmptyString",
			URL:    "",
			Result: "",
		},
	}

	for _, tCase := range tMatrix {
		t.Run(tCase.Name, func(t *testing.T) {
			assert.Equal(t, tCase.Result, extractRepositoryName(tCase.URL), "Should return expected result")
		})
	}

}

func TestCloneRepository(t *testing.T) {
	t.Run("NoDirectory", func(t *testing.T) {
		require := require.New(t)

		oldDir, err := os.Getwd()
		require.NoError(err, "Should get current working directory")

		dir := t.TempDir()
		require.NoError(os.Chdir(dir), "Should change working directory")
		t.Cleanup(func() {
			require.NoError(os.Chdir(oldDir), "Should change back to original working directory")
		})

		err = CloneRepository("https://github.com/heathcliff26/gh-utility.git", "", 1)
		require.NoError(err, "Should clone repository")

		ls, err := os.ReadDir(dir)
		require.NoError(err, "Should read directory")
		require.Len(ls, 1, "Should have one directory")

		entry := ls[0]
		require.True(entry.IsDir(), "Should be a directory")
		require.Equal("gh-utility", entry.Name(), "Should have expected directory name")
	})
	t.Run("WithDirectory", func(t *testing.T) {
		require := require.New(t)

		dir := t.TempDir()

		err := CloneRepository("https://github.com/heathcliff26/gh-utility.git", dir, 1)
		require.NoError(err, "Should clone repository")

		ls, err := os.ReadDir(dir)
		require.NoError(err, "Should read directory")
		require.GreaterOrEqual(len(ls), 1, "Should have at least one entry")
	})
	t.Run("InvalidURL", func(t *testing.T) {
		require := require.New(t)

		dir := t.TempDir()

		err := CloneRepository("not-a-valid-url", dir, 1)
		require.ErrorContains(err, "repository not found", "Should not clone repository")
	})
}
