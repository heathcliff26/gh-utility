package git

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/heathcliff26/gh-utility/testutils"
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
			Name:   "NoOwner",
			URL:    "https://github.com/gh-utility",
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

func TestExtractRepositoryOwnerAndName(t *testing.T) {
	tMatrix := []struct {
		Name   string
		URL    string
		Result string
	}{
		{
			Name:   "HTTP",
			URL:    "https://github.com/heathcliff26/gh-utility.git",
			Result: "heathcliff26/gh-utility",
		},
		{
			Name:   "SSH",
			URL:    "git@github.com:heathcliff26/gh-utility.git",
			Result: "heathcliff26/gh-utility",
		},
		{
			Name:   "NoGitSuffix",
			URL:    "https://github.com/heathcliff26/gh-utility",
			Result: "heathcliff26/gh-utility",
		},
		{
			Name:   "EmptyString",
			URL:    "",
			Result: "",
		},
	}

	for _, tCase := range tMatrix {
		t.Run(tCase.Name, func(t *testing.T) {
			assert.Equal(t, tCase.Result, extractRepositoryOwnerAndName(tCase.URL), "Should return expected result")
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

func TestOpenRepository(t *testing.T) {
	t.Run("InvalidRepository", func(t *testing.T) {
		assert := assert.New(t)

		r, err := OpenRepository("invalid-path")
		assert.Nil(r, "Should not return a repository")
		assert.Error(err, "Should return an error")
	})
	t.Run("Success", func(t *testing.T) {
		assert := assert.New(t)

		dir, _ := testutils.NewTestRepository(t)

		r, err := OpenRepository(dir)

		assert.NoError(err, "Should not return an error")
		assert.NotNil(r, "Should return the same repository")
	})
}

func TestGetChangedFiles(t *testing.T) {
	require := require.New(t)

	dir, repo := testutils.NewTestRepository(t)
	r := &Repository{repo: repo}

	err := os.Remove(filepath.Join(dir, testutils.DeletedFile))
	require.NoError(err, "Should delete file")
	err = os.WriteFile(filepath.Join(dir, testutils.ChangedFile), []byte("This file has been changed"), 0644)
	require.NoError(err, "Should modify file")
	err = os.WriteFile(filepath.Join(dir, testutils.StagedFile), []byte("This file has been changed"), 0644)
	require.NoError(err, "Should modify file")
	err = os.WriteFile(filepath.Join(dir, "new.txt"), []byte("This is new"), 0644)
	require.NoError(err, "Should create file")

	tree, err := r.repo.Worktree()
	require.NoError(err, "Should return worktree")
	_, err = tree.Add(testutils.StagedFile)
	require.NoError(err, "Should stage file")

	expectedResult := []string{testutils.DeletedFile, testutils.ChangedFile, testutils.StagedFile, "new.txt"}
	result, err := r.GetChangedFiles()
	require.NoError(err, "Should return changed files")
	require.ElementsMatch(expectedResult, result, "Should return the expected files")
}

func TestGetTreeHash(t *testing.T) {
	t.Run("NoHead", func(t *testing.T) {
		assert := assert.New(t)

		dir, repo := testutils.NewTestRepository(t)
		r := &Repository{repo: repo}

		assert.NoError(os.RemoveAll(filepath.Join(dir, ".git")), "Should remove .git repo")

		hash, err := r.GetTreeHash()
		assert.Error(err, "Should return an error")
		assert.Empty(hash, "Should not return a hash")
	})
	t.Run("Success", func(t *testing.T) {
		assert := assert.New(t)

		_, repo := testutils.NewTestRepository(t)
		r := &Repository{repo: repo}

		hash, err := r.GetTreeHash()
		assert.NoError(err)
		assert.NotEmpty(hash, "Should return a hash")
	})
}

func TestCommitHash(t *testing.T) {
	t.Run("NoHead", func(t *testing.T) {
		assert := assert.New(t)

		dir, repo := testutils.NewTestRepository(t)
		r := &Repository{repo: repo}

		assert.NoError(os.RemoveAll(filepath.Join(dir, ".git")), "Should remove .git repo")

		hash, err := r.GetCommitHash()
		assert.Error(err, "Should return an error")
		assert.Empty(hash, "Should not return a hash")
	})
	t.Run("Success", func(t *testing.T) {
		assert := assert.New(t)

		_, repo := testutils.NewTestRepository(t)
		r := &Repository{repo: repo}

		hash, err := r.GetCommitHash()
		assert.NoError(err)
		assert.NotEmpty(hash, "Should return a hash")
	})
}

func TestGetRemote(t *testing.T) {
	t.Run("NoHead", func(t *testing.T) {
		assert := assert.New(t)

		dir, repo := testutils.NewTestRepository(t)
		r := &Repository{repo: repo}

		assert.NoError(os.RemoveAll(filepath.Join(dir, ".git")), "Should remove .git repo")

		remote, err := r.GetRemote()
		assert.Error(err, "Should return an error")
		assert.Empty(remote, "Should not return a remote")
	})
	t.Run("NoRemotes", func(t *testing.T) {
		assert := assert.New(t)

		_, repo := testutils.NewTestRepository(t)
		r := &Repository{repo: repo}

		remote, err := r.GetRemote()
		assert.ErrorContains(err, "no remotes found")
		assert.Empty(remote, "Should not return a remote")
	})
	t.Run("Success", func(t *testing.T) {
		assert := assert.New(t)

		_, repo := testutils.NewTestRepository(t)
		r := &Repository{repo: repo}

		rc := config.RemoteConfig{
			Name: "origin",
			URLs: []string{"https://github.com/heathcliff26/gh-utility.git"},
		}
		_, err := r.repo.CreateRemote(&rc)
		assert.NoError(err, "Should create remote for testing")

		remote, err := r.GetRemote()
		assert.NoError(err, "Should succeed")
		assert.Equal("heathcliff26/gh-utility", remote, "Should not return a remote")
	})
}
