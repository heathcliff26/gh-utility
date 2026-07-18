package git

import (
	"fmt"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/heathcliff26/gh-utility/pkg/utils"
)

// Clone the given repository into the specified directory.
// If the directory is empty, the repository name will be used as the directory name.
func CloneRepository(repoURL string, directory string, fetchDepth int) error {
	opt := newCloneOptions(repoURL, fetchDepth)

	if directory == "" {
		directory = extractRepositoryName(repoURL)
	}

	_, err := git.PlainClone(directory, false, opt)
	return err
}

func newCloneOptions(repoURL string, fetchDepth int) *git.CloneOptions {
	opt := &git.CloneOptions{
		URL:   repoURL,
		Depth: fetchDepth,
	}

	token := utils.GetToken()
	if token != "" {
		opt.Auth = &http.BasicAuth{
			Username: "x-access-token",
			Password: token,
		}
	}

	return opt
}

func extractRepositoryName(url string) string {
	url = extractRepositoryOwnerAndName(url)
	s := strings.Split(url, "/")
	return s[len(s)-1]
}

func extractRepositoryOwnerAndName(url string) string {
	s := strings.Split(url, ":")
	s = strings.Split(s[len(s)-1], "/")
	if len(s) == 1 {
		url = s[0]
	} else {
		url = s[len(s)-2] + "/" + s[len(s)-1]
	}
	return strings.TrimSuffix(url, ".git")
}

type Repository struct {
	repo *git.Repository
}

func OpenRepository(path string) (*Repository, error) {
	repo, err := git.PlainOpen(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open repository: %w", err)
	}
	return &Repository{
		repo: repo,
	}, nil
}

// Scan the current worktree and return all changed files.
func (r *Repository) GetChangedFiles() ([]string, error) {
	tree, err := r.repo.Worktree()
	if err != nil {
		return nil, fmt.Errorf("failed to read repository worktree: %w", err)
	}
	status, err := tree.Status()
	if err != nil {
		return nil, fmt.Errorf("failed to read worktree status: %w", err)
	}
	result := make([]string, 0, 10)
	for file, fileStatus := range status {
		if fileStatus.Worktree != git.Unmodified || fileStatus.Staging != git.Unmodified {
			result = append(result, file)
		}
	}
	return result, nil
}

// Return the Tree Hash of the current commit in the repository.
func (r *Repository) GetTreeHash() (string, error) {
	headRef, err := r.repo.Head()
	if err != nil {
		return "", fmt.Errorf("failed to get HEAD reference: %w", err)
	}
	commit, err := r.repo.CommitObject(headRef.Hash())
	if err != nil {
		return "", fmt.Errorf("failed to get current commit object: %w", err)
	}
	return commit.TreeHash.String(), nil
}

// Return the Hash of the current commit in the repository.
func (r *Repository) GetCommitHash() (string, error) {
	headRef, err := r.repo.Head()
	if err != nil {
		return "", fmt.Errorf("failed to get HEAD reference: %w", err)
	}
	return headRef.Hash().String(), nil
}

// Return the remote name.
func (r *Repository) GetRemote() (string, error) {
	remotes, err := r.repo.Remotes()
	if err != nil {
		return "", fmt.Errorf("failed to get remotes: %w", err)
	}
	if len(remotes) == 0 {
		return "", fmt.Errorf("no remotes found")
	}

	url := remotes[0].Config().URLs[0]
	return extractRepositoryOwnerAndName(url), nil
}

// Return the name of the current branch
func (r *Repository) CurrentBranch() (string, error) {
	head, err := r.repo.Head()
	if err != nil {
		return "", fmt.Errorf("failed to get HEAD reference: %w", err)
	}
	ref := head.Name().String()
	return strings.TrimPrefix(ref, "refs/heads/"), nil
}
