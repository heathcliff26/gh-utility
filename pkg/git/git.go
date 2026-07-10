package git

import (
	"os"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

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

	token := os.Getenv("GITHUB_TOKEN")
	if token != "" {
		opt.Auth = &http.BasicAuth{
			Username: "x-access-token",
			Password: token,
		}
	}

	return opt
}

func extractRepositoryName(repoURL string) string {
	s := strings.Split(repoURL, ":")
	s = strings.Split(s[len(s)-1], "/")
	name := s[len(s)-1]
	name = strings.TrimSuffix(name, ".git")
	return name
}
