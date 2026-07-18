package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	DefaultEndpoint = "https://api.github.com"

	defaultClientTimeout = 10 * time.Second
)

type Client struct {
	httpClient *http.Client
	endpoint   string
}

// Create a new client
func NewClient(endpoint string) *Client {
	httpClient := &http.Client{
		Timeout: defaultClientTimeout,
	}

	return &Client{
		httpClient: httpClient,
		endpoint:   endpoint,
	}
}

// Get an installations token for the GitHub app
// API endpoint: POST /app/installations/{installation_id}/access_tokens
func (c *Client) GetToken(keyPath string, clientID string, installationID string) (string, error) {
	// #nosec G304 -- Filepath is decided by the user.
	bytes, err := os.ReadFile(keyPath)
	if err != nil {
		return "", fmt.Errorf("failed to read keyfile: %w", err)
	}

	// Need to replace literal `\n` with actual newlines.
	// Otherwise it is harder to use this in environments where the key might be passed as a single line string.
	keyStr := strings.ReplaceAll(string(bytes), `\n`, "\n")
	key, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(keyStr))
	if err != nil {
		return "", fmt.Errorf("failed to parse keyfile: %w", err)
	}

	claims := &jwt.RegisteredClaims{
		IssuedAt:  jwt.NewNumericDate(time.Now().Add(-time.Minute)),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(9 * time.Minute)),
		Issuer:    clientID,
	}
	jwToken := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	signedJWToken, err := jwToken.SignedString(key)
	if err != nil {
		return "", fmt.Errorf("failed to sign JSONWebToken: %w", err)
	}

	req, err := newRequest(http.MethodPost, fmt.Sprintf("%s/app/installations/%s/access_tokens", c.endpoint, installationID), nil, signedJWToken)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	var tokenResp TokenResponse
	err = c.do(req, http.StatusCreated, &tokenResp)
	if err != nil {
		return "", err
	}
	return tokenResp.Token, nil
}

// Create a new Git tree object with the give files.
// API endpoint: POST /repos/{owner}/{repo}/git/trees
// Parameters:
// - token: The GitHub app installation token
// - repo: The repository name in the format "owner/repo".
// - files: A list of file paths
// - baseTree: The SHA of the base tree to create the new tree from.
// Returns:
// - The SHA of the created tree object
// - An error if any occurred
func (c *Client) CreateTree(token string, repo string, files []string, baseTree string) (string, error) {
	objects := make([]*TreeObject, len(files))
	for i, file := range files {
		obj, err := fileToTreeObject(file)
		if err != nil {
			return "", fmt.Errorf("failed to create tree object for file '%s': %w", file, err)
		}
		objects[i] = obj
	}

	treeReq := &TreeRequest{
		Tree:     objects,
		BaseTree: baseTree,
	}

	req, err := newRequest(http.MethodPost, fmt.Sprintf("%s/repos/%s/git/trees", c.endpoint, repo), treeReq, token)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	var res TreeResponse
	err = c.do(req, http.StatusCreated, &res)
	if err != nil {
		return "", err
	}
	return res.SHA, nil
}

// Create a new Git commit object
// API endpoint: POST /repos/{owner}/{repo}/git/commits
// Parameters:
// - token: The GitHub app installation token
// - repo: The repository name in the format "owner/repo".
// - msg: Commit message
// - tree: The SHA of the tree the commit references
// - parents: List of parent commit SHA
// Returns:
// - The SHA of the created commit
// - An error if any occurred
func (c *Client) CreateCommit(token string, repo string, msg string, tree string, parents []string) (string, error) {
	commit := &CommitRequest{
		Message: msg,
		Tree:    tree,
		Parents: parents,
	}

	req, err := newRequest(http.MethodPost, fmt.Sprintf("%s/repos/%s/git/commits", c.endpoint, repo), commit, token)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	var res CommitResponse
	err = c.do(req, http.StatusCreated, &res)
	if err != nil {
		return "", err
	}
	return res.SHA, nil
}

// Check if the given branch exists
// API endpoint: POST /repos/{owner}/{repo}/git/refs/{ref}
// Parameters:
// - token: The GitHub app installation token
// - repo: The repository name in the format "owner/repo".
// - name: Name of the branch
// Returns:
// - If the branch exists
// - An error if any occurred
func (c *Client) ExistsBranch(token string, repo string, name string) (bool, error) {
	req, err := newRequest(http.MethodGet, fmt.Sprintf("%s/repos/%s/git/refs/heads/%s", c.endpoint, repo, name), nil, token)
	if err != nil {
		return false, fmt.Errorf("failed to create request: %w", err)
	}

	res, err := c.httpClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("failed to send request: %w", err)
	}
	defer res.Body.Close()

	switch res.StatusCode {
	case http.StatusOK:
		return true, nil
	case http.StatusNotFound:
		return false, nil
	}
	return false, fmt.Errorf("received unexpected response code: %d", res.StatusCode)
}

// Create a new branch at the given commit
// API endpoint: POST /repos/{owner}/{repo}/git/refs
// Parameters:
// - token: The GitHub app installation token
// - repo: The repository name in the format "owner/repo".
// - name: Name of the branch
// - commit: SHA of the commit
// Returns:
// - An error if any occurred
func (c *Client) CreateBranch(token string, repo string, name string, commit string) error {
	branch := &BranchRequest{
		Ref: "refs/heads/" + name,
		SHA: commit,
	}

	req, err := newRequest(http.MethodPost, fmt.Sprintf("%s/repos/%s/git/refs", c.endpoint, repo), branch, token)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	err = c.do(req, http.StatusCreated, nil)
	if err != nil {
		return fmt.Errorf("failed to create branch: %w", err)
	}
	return nil
}

// Update the branch to the given commit. Will force push
// API endpoint: POST /repos/{owner}/{repo}/git/refs/{ref}
// Parameters:
// - token: The GitHub app installation token
// - repo: The repository name in the format "owner/repo".
// - name: Name of the branch
// - commit: SHA of the commit
// Returns:
// - An error if any occurred
func (c *Client) UpdateBranch(token string, repo string, name string, commit string) error {
	force := true
	branch := &BranchRequest{
		SHA:   commit,
		Force: &force,
	}

	req, err := newRequest(http.MethodPatch, fmt.Sprintf("%s/repos/%s/git/refs/heads/%s", c.endpoint, repo, name), branch, token)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	err = c.do(req, http.StatusOK, nil)
	if err != nil {
		return fmt.Errorf("failed to update branch: %w", err)
	}
	return nil
}

// Check if the branch exists and call Create or Update afterwards.
func (c *Client) CreateOrUpdateBranch(token string, repo string, name string, commit string) error {
	exists, err := c.ExistsBranch(token, repo, name)
	if err != nil {
		return fmt.Errorf("failed to check if branch exists: %w", err)
	}
	if exists {
		return c.UpdateBranch(token, repo, name, commit)
	} else {
		return c.CreateBranch(token, repo, name, commit)
	}
}

// Return a PR for the given branch
// API endpoint: GET /repos/{owner}/{repo}/pulls
// Parameters:
// - token: The GitHub app installation token
// - repo: The repository name in the format "owner/repo".
// - branch: Name of the branch
// Returns:
// - The Pull Request or nil if none exists
// - An error if any occurred
func (c *Client) GetPullRequestForBranch(token, repo, branch string) (*PrResponse, error) {
	owner := strings.Split(repo, "/")[0]
	query := url.QueryEscape(fmt.Sprintf("%s:%s", owner, branch))

	req, err := newRequest(http.MethodGet, fmt.Sprintf("%s/repos/%s/pulls?head=%s", c.endpoint, repo, query), nil, token)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	var res []PrResponse
	err = c.do(req, http.StatusOK, &res)
	if err != nil {
		return nil, fmt.Errorf("failed to list pull requests: %w", err)
	}

	for _, pr := range res {
		if pr.Head.Ref == branch {
			return &pr, nil
		}
	}
	return nil, nil
}

// Create a pull request with the given title and body
// API endpoint: POST /repos/{owner}/{repo}/pulls
// Parameters:
// - token: The GitHub app installation token
// - repo: The repository name in the format "owner/repo".
// - title: Title of the PR
// - body: Body of the PR
// - head: Head branch of the PR
// - base: Base branch of the PR
// Returns:
// - Pull Request
// - An error if any occurred
func (c *Client) CreatePullRequest(token, repo, title, body, head, base string) (*PrResponse, error) {
	owner := strings.Split(repo, "/")[0]
	pr := &PrRequest{
		Title: title,
		Head:  fmt.Sprintf("%s:%s", owner, head),
		Base:  base,
		Body:  body,
	}

	req, err := newRequest(http.MethodPost, fmt.Sprintf("%s/repos/%s/pulls", c.endpoint, repo), pr, token)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	var res PrResponse
	err = c.do(req, http.StatusCreated, &res)
	if err != nil {
		return nil, fmt.Errorf("failed to create pull request: %w", err)
	}
	return &res, nil
}

// Update the pull request with the given title and body
// API endpoint: PATCH /repos/{owner}/{repo}/pulls/{pull_number}
// Parameters:
// - token: The GitHub app installation token
// - repo: The repository name in the format "owner/repo".
// - number: Number of the PR to update
// - title: Title of the PR
// - body: Body of the PR
// - base: Base branch of the PR
// Returns:
// - Pull Request
// - An error if any occurred
func (c *Client) UpdatePullRequest(token, repo string, number int, title, body, base string) (*PrResponse, error) {
	pr := &PrRequest{
		Title: title,
		Base:  base,
		Body:  body,
	}

	req, err := newRequest(http.MethodPatch, fmt.Sprintf("%s/repos/%s/pulls/%d", c.endpoint, repo, number), pr, token)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	var res PrResponse
	err = c.do(req, http.StatusOK, &res)
	if err != nil {
		return nil, fmt.Errorf("failed to update pull request: %w", err)
	}
	return &res, nil
}

// Check if there is already an existing PR for the head branch.
// Call Create/UpdatePullRequest as needed.
// Will check if there are changes before calling update.
func (c *Client) CreateOrUpdatePullRequest(token, repo, title, body, head, base string) (*PrResponse, error) {
	pr, err := c.GetPullRequestForBranch(token, repo, head)
	if err != nil {
		return nil, fmt.Errorf("failed to check for existing PR: %w", err)
	}
	if pr == nil {
		return c.CreatePullRequest(token, repo, title, body, head, base)
	}

	if pr.Title == title && pr.Body == body && pr.Base.Ref == base {
		return pr, nil
	}

	return c.UpdatePullRequest(token, repo, pr.Number, title, body, base)
}

// Send the given http request and parse the returned result
func (c *Client) do(req *http.Request, status int, v any) error {
	// #nosec G704 -- Endpoint should be user controlled, actual Paths are hardcoded
	res, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if res.StatusCode != status {
		return fmt.Errorf("request returned %d, expected %d: %s", res.StatusCode, status, string(body))
	}
	if v != nil {
		err = json.Unmarshal(body, v)
		if err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}
	return nil
}
