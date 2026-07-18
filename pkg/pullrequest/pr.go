package pullrequest

import (
	"fmt"

	"github.com/heathcliff26/gh-utility/pkg/client"
	"github.com/heathcliff26/gh-utility/pkg/git"
)

// Commit the current changes to the given branch.
// Will force push the branch if needed.
// Returns the hashes of the created tree and commit
func Commit(c *client.Client, dir, msg, branch, token string) (string, string, error) {
	repo, err := git.OpenRepository(dir)
	if err != nil {
		return "", "", err
	}
	treeHash, err := repo.GetTreeHash()
	if err != nil {
		return "", "", err
	}
	changedFiles, err := repo.GetChangedFiles()
	if err != nil {
		return "", "", err
	}

	if len(changedFiles) == 0 {
		return "", "", nil
	}

	remote, err := repo.GetRemote()
	if err != nil {
		return "", "", err
	}

	treeHash, err = c.CreateTree(token, remote, changedFiles, treeHash)
	if err != nil {
		return "", "", fmt.Errorf("failed to create tree: %w", err)
	}

	commitHash, err := repo.GetCommitHash()
	if err != nil {
		return treeHash, "", err
	}

	commitHash, err = c.CreateCommit(token, remote, msg, treeHash, []string{commitHash})
	if err != nil {
		return treeHash, "", fmt.Errorf("failed to create commit: %w", err)
	}

	err = c.CreateOrUpdateBranch(token, remote, branch, commitHash)
	if err != nil {
		return treeHash, commitHash, fmt.Errorf("failed to create/update branch: %w", err)
	}

	return treeHash, commitHash, nil
}

// Check if a pull request already exists.
// Create if not exists, update if exists.
// Returns a url to the pull request.
func PullRequest(c *client.Client, dir, branch, title, body, token string, labels []string) (string, error) {
	repo, err := git.OpenRepository(dir)
	if err != nil {
		return "", err
	}
	remote, err := repo.GetRemote()
	if err != nil {
		return "", err
	}
	base, err := repo.CurrentBranch()
	if err != nil {
		return "", err
	}

	pr, err := c.CreateOrUpdatePullRequest(token, remote, title, body, branch, base)
	if err != nil {
		return "", err
	}
	return pr.HtmlUrl, nil
}
