package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/heathcliff26/gh-utility/testutils"
	"github.com/stretchr/testify/assert"
)

func TestNewClient(t *testing.T) {
	assert := assert.New(t)

	client := NewClient(DefaultEndpoint)
	assert.NotNil(client)
	assert.Equal(DefaultEndpoint, client.endpoint)
	assert.NotNil(client.httpClient)
	assert.Equal(defaultClientTimeout, client.httpClient.Timeout)
}

func TestGetToken(t *testing.T) {
	t.Run("MissingKey", func(t *testing.T) {
		assert := assert.New(t)

		client := NewClient("http://localhost")
		token, err := client.GetToken("nothing", "abcdf", "12345")
		assert.ErrorContains(err, "failed to read keyfile", "Should return error")
		assert.Empty(token)
	})
	t.Run("MalformedKey", func(t *testing.T) {
		assert := assert.New(t)

		client := NewClient("http://localhost")
		token, err := client.GetToken("testdata/fake-key.txt", "abcdf", "12345")
		assert.ErrorContains(err, "failed to parse keyfile", "Should return error")
		assert.Empty(token)
	})
	t.Run("FailedRequest", func(t *testing.T) {
		assert := assert.New(t)

		client := NewClient("http://localhost:6666")

		privateKey := testutils.GenerateRSAKey(t)

		token, err := client.GetToken(privateKey, "abcdf", "12345")
		assert.ErrorContains(err, "failed to send request", "Should return error")
		assert.Empty(token)
	})
	t.Run("RequestReturnsFailure", func(t *testing.T) {
		assert := assert.New(t)

		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer s.Close()

		client := NewClient(s.URL)

		privateKey := testutils.GenerateRSAKey(t)

		token, err := client.GetToken(privateKey, "abcdf", "12345")
		assert.ErrorContains(err, "request returned 500, expected 201:", "Should return error")
		assert.Empty(token)
	})
	t.Run("ParseResponseFailure", func(t *testing.T) {
		assert := assert.New(t)

		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte("not json"))
		}))
		defer s.Close()

		client := NewClient(s.URL)

		privateKey := testutils.GenerateRSAKey(t)

		token, err := client.GetToken(privateKey, "abcdf", "12345")
		assert.ErrorContains(err, "failed to decode response", "Should return error")
		assert.Empty(token)
	})
	t.Run("Success", func(t *testing.T) {
		assert := assert.New(t)

		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(http.MethodPost, r.Method, "Should use POST method")
			assert.Equal("/app/installations/12345/access_tokens", r.URL.Path, "Should use correct path")
			assert.NotEmpty(r.Header.Get("Accept"), "Should set Accept header")

			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"token": "abc", "expires_at": "future"}`))
		}))
		defer s.Close()

		client := NewClient(s.URL)

		privateKey := testutils.GenerateRSAKey(t)

		token, err := client.GetToken(privateKey, "abcdf", "12345")
		assert.NoError(err, "Should succeed")
		assert.Equal("abc", token, "Should return token")
	})
}

func TestCreateTree(t *testing.T) {
	token, repo, baseTree := "testtoken", "test/repo", "base"
	files := []string{"testdata/tree/deleted.txt"}

	newServer := func(t *testing.T, status int) *httptest.Server {
		return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert := assert.New(t)

			assert.Equal(http.MethodPost, r.Method, "Should use POST method")
			assert.Equal("/repos/test/repo/git/trees", r.URL.Path, "Should use correct path")
			assert.NotEmpty(r.Header.Get("Accept"), "Should set Accept header")

			var treeReq TreeRequest
			err := json.NewDecoder(r.Body).Decode(&treeReq)
			assert.NoError(err, "Should decode request")

			assert.Equal(baseTree, treeReq.BaseTree, "Should have correct base tree")
			assert.Len(treeReq.Tree, 1, "Should have 1 tree object")

			w.WriteHeader(status)
			_, _ = w.Write([]byte(`{"sha": "abc"}`))
		}))
	}

	t.Run("Success", func(t *testing.T) {
		assert := assert.New(t)

		s := newServer(t, http.StatusCreated)
		defer s.Close()

		client := NewClient(s.URL)

		hash, err := client.CreateTree(token, repo, files, baseTree, "")
		assert.NoError(err, "Should create tree")
		assert.Equal("abc", hash, "Should return new hash")
	})
	t.Run("Failure", func(t *testing.T) {
		assert := assert.New(t)

		s := newServer(t, http.StatusInternalServerError)
		defer s.Close()

		client := NewClient(s.URL)

		hash, err := client.CreateTree(token, repo, files, baseTree, "")
		assert.Error(err, "Should fail")
		assert.Empty(hash, "Should not return a hash")
	})
}

func TestCreateCommit(t *testing.T) {
	token, repo, msg, tree, parent := "testtoken", "test/repo", "Test Commit", "abcdef", "123456"

	newServer := func(t *testing.T, status int) *httptest.Server {
		return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert := assert.New(t)

			assert.Equal(http.MethodPost, r.Method, "Should use POST method")
			assert.Equal("/repos/test/repo/git/commits", r.URL.Path, "Should use correct path")
			assert.NotEmpty(r.Header.Get("Accept"), "Should set Accept header")

			var commitReq CommitRequest
			err := json.NewDecoder(r.Body).Decode(&commitReq)
			assert.NoError(err, "Should decode request")

			assert.Equal(msg, commitReq.Message, "Should have correct message")
			assert.Equal(tree, commitReq.Tree, "Should have tree ref")
			assert.Equal(parent, commitReq.Parents[0], "Should have parent ref")

			w.WriteHeader(status)
			_, _ = w.Write([]byte(`{"sha": "abc"}`))
		}))
	}

	t.Run("Success", func(t *testing.T) {
		assert := assert.New(t)

		s := newServer(t, http.StatusCreated)
		defer s.Close()

		client := NewClient(s.URL)

		hash, err := client.CreateCommit(token, repo, msg, tree, []string{parent})
		assert.NoError(err, "Should create commit")
		assert.Equal("abc", hash, "Should return new hash")
	})
	t.Run("Failure", func(t *testing.T) {
		assert := assert.New(t)

		s := newServer(t, http.StatusInternalServerError)
		defer s.Close()

		client := NewClient(s.URL)

		hash, err := client.CreateCommit(token, repo, msg, tree, []string{parent})
		assert.Error(err, "Should fail")
		assert.Empty(hash, "Should not return a hash")
	})
}

func TestCreateOrUpdateBranch(t *testing.T) {
	var statusCodes, expectedEndpoints []int
	i := 0

	token, repo, branch, commit := "testtoken", "test/repo", "testbranch", "abcdef"

	newServer := func(t *testing.T) *httptest.Server {
		return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert := assert.New(t)

			assert.NotEmpty(r.Header.Get("Accept"), "Should set Accept header")

			var branchReq BranchRequest

			switch expectedEndpoints[i] {
			case 0:
				assert.Equal(http.MethodGet, r.Method, "Should use GET method")
				assert.Equal("/repos/test/repo/git/refs/heads/testbranch", r.URL.Path, "Should use correct path")
			case 1:
				assert.Equal(http.MethodPost, r.Method, "Should use POST method")
				assert.Equal("/repos/test/repo/git/refs", r.URL.Path, "Should use correct path")

				err := json.NewDecoder(r.Body).Decode(&branchReq)
				assert.NoError(err, "Should decode request")
				assert.Equal(commit, branchReq.SHA, "Should have commit SHA")
				assert.Equal("refs/heads/"+branch, branchReq.Ref, "Should have branch")
				assert.Nil(branchReq.Force, "Should not have force set")
			case 2:
				assert.Equal(http.MethodPatch, r.Method, "Should use PATCH method")
				assert.Equal("/repos/test/repo/git/refs/heads/testbranch", r.URL.Path, "Should use correct path")

				err := json.NewDecoder(r.Body).Decode(&branchReq)
				assert.NoError(err, "Should decode request")
				assert.Equal(commit, branchReq.SHA, "Should have commit SHA")
				assert.Empty(branchReq.Ref, "Should not have branch")
				assert.NotNil(branchReq.Force, "Should have force set")
			default:
				t.Fatal("Unexpected endpoint")
			}

			w.WriteHeader(statusCodes[i])
			i++
		}))
	}

	t.Run("ExistsFails", func(t *testing.T) {
		assert := assert.New(t)

		statusCodes = []int{http.StatusInternalServerError}
		expectedEndpoints = []int{0}
		i = 0

		s := newServer(t)
		defer s.Close()

		client := NewClient(s.URL)

		err := client.CreateOrUpdateBranch(token, repo, branch, commit)
		assert.ErrorContains(err, "failed to check if branch exists: ", "Should fail with correct error")

		assert.Equal(1, i, "Should have called the endpoint the correct number of times")
	})
	t.Run("CreateBranch", func(t *testing.T) {
		assert := assert.New(t)

		statusCodes = []int{http.StatusNotFound, http.StatusCreated}
		expectedEndpoints = []int{0, 1}
		i = 0

		s := newServer(t)
		defer s.Close()

		client := NewClient(s.URL)

		err := client.CreateOrUpdateBranch(token, repo, branch, commit)
		assert.NoError(err, "Should succeed")

		assert.Equal(2, i, "Should have called the endpoint the correct number of times")
	})
	t.Run("UpdateBranch", func(t *testing.T) {
		assert := assert.New(t)

		statusCodes = []int{http.StatusOK, http.StatusOK}
		expectedEndpoints = []int{0, 2}
		i = 0

		s := newServer(t)
		defer s.Close()

		client := NewClient(s.URL)

		err := client.CreateOrUpdateBranch(token, repo, branch, commit)
		assert.NoError(err, "Should succeed")

		assert.Equal(2, i, "Should have called the endpoint the correct number of times")
	})
	t.Run("CreateBranchFail", func(t *testing.T) {
		assert := assert.New(t)

		statusCodes = []int{http.StatusNotFound, http.StatusInternalServerError}
		expectedEndpoints = []int{0, 1}
		i = 0

		s := newServer(t)
		defer s.Close()

		client := NewClient(s.URL)

		err := client.CreateOrUpdateBranch(token, repo, branch, commit)
		assert.ErrorContains(err, "failed to create branch:", "Should fail with correct error")

		assert.Equal(2, i, "Should have called the endpoint the correct number of times")
	})
	t.Run("UpdateBranchFail", func(t *testing.T) {
		assert := assert.New(t)

		statusCodes = []int{http.StatusOK, http.StatusInternalServerError}
		expectedEndpoints = []int{0, 2}
		i = 0

		s := newServer(t)
		defer s.Close()

		client := NewClient(s.URL)

		err := client.CreateOrUpdateBranch(token, repo, branch, commit)
		assert.ErrorContains(err, "failed to update branch:", "Should fail with correct error")

		assert.Equal(2, i, "Should have called the endpoint the correct number of times")
	})
}

func TestGetPullRequestForBranch(t *testing.T) {
	branch := "testbranch"
	correctPR, wrongPR1, wrongPR2 := PrResponse{}, PrResponse{}, PrResponse{}
	correctPR.Head.Ref = branch
	wrongPR1.Head.Ref = "wrong-branch-1"
	wrongPR2.Head.Ref = "wrong-branch-2"

	tMatrix := []struct {
		Name   string
		PRs    []PrResponse
		Result *PrResponse
	}{
		{
			Name:   "NoPRs",
			PRs:    []PrResponse{},
			Result: nil,
		},
		{
			Name:   "MultiplePRs",
			PRs:    []PrResponse{wrongPR1, correctPR, wrongPR2},
			Result: &correctPR,
		},
		{
			Name:   "OnlyWrongPRs",
			PRs:    []PrResponse{wrongPR1, wrongPR2},
			Result: nil,
		},
		{
			Name:   "SingleHit",
			PRs:    []PrResponse{correctPR},
			Result: &correctPR,
		},
	}

	for _, tCase := range tMatrix {
		t.Run(tCase.Name, func(t *testing.T) {
			assert := assert.New(t)

			s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(http.MethodGet, r.Method, "Should use GET method")
				assert.Equal("/repos/test/repo/pulls", r.URL.Path, "Should use correct path")
				assert.NotEmpty(r.Header.Get("Accept"), "Should set Accept header")

				err := json.NewEncoder(w).Encode(tCase.PRs)
				assert.NoError(err, "Should send response")
			}))
			defer s.Close()

			client := NewClient(s.URL)

			pr, err := client.GetPullRequestForBranch("testtoken", "test/repo", branch)
			assert.Equal(tCase.Result, pr, "Should return correct result")
			assert.NoError(err, "Should not fail")
		})
	}
	t.Run("FailedRequest", func(t *testing.T) {
		assert := assert.New(t)

		client := NewClient("http://localhost:6666")

		pr, err := client.GetPullRequestForBranch("testtoken", "test/repo", branch)
		assert.Nil(pr, "Should not return a result")
		assert.Error(err, "Should fail")
	})
}

func TestCreatePullRequest(t *testing.T) {
	token, repo, title, body, head, base := "testtoken", "test/repo", "Test Title", "Test Body", "testhead", "testbase"

	newServer := func(t *testing.T, status int) *httptest.Server {
		return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert := assert.New(t)

			assert.Equal(http.MethodPost, r.Method, "Should use POST method")
			assert.Equal("/repos/test/repo/pulls", r.URL.Path, "Should use correct path")
			assert.NotEmpty(r.Header.Get("Accept"), "Should set Accept header")

			var req PrRequest
			err := json.NewDecoder(r.Body).Decode(&req)
			assert.NoError(err, "Should decode request")

			assert.Equal(title, req.Title, "Should have correct title")
			assert.Equal(body, req.Body, "Should have correct body")
			assert.Equal("test:"+head, req.Head, "Should have correct head")
			assert.Equal(base, req.Base, "Should have correct base")

			w.WriteHeader(status)
			_, _ = w.Write([]byte(`{"html_url": "testurl"}`))
		}))
	}

	t.Run("Success", func(t *testing.T) {
		assert := assert.New(t)

		s := newServer(t, http.StatusCreated)
		defer s.Close()

		client := NewClient(s.URL)

		res, err := client.CreatePullRequest(token, repo, title, body, head, base)
		assert.NoError(err, "Should create PR")
		assert.Equal("testurl", res.HtmlUrl, "Should return correct PR")
	})
	t.Run("Failure", func(t *testing.T) {
		assert := assert.New(t)

		s := newServer(t, http.StatusInternalServerError)
		defer s.Close()

		client := NewClient(s.URL)

		res, err := client.CreatePullRequest(token, repo, title, body, head, base)
		assert.Error(err, "Should fail")
		assert.Nil(res, "Should not return a PR")
	})
}

func TestUpdatePullRequest(t *testing.T) {
	token, repo, title, body, base := "testtoken", "test/repo", "Test Title", "Test Body", "testbase"
	number := 1234

	newServer := func(t *testing.T, status int) *httptest.Server {
		return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert := assert.New(t)

			assert.Equal(http.MethodPatch, r.Method, "Should use PATCH method")
			assert.Equal("/repos/test/repo/pulls/1234", r.URL.Path, "Should use correct path")
			assert.NotEmpty(r.Header.Get("Accept"), "Should set Accept header")

			var req PrRequest
			err := json.NewDecoder(r.Body).Decode(&req)
			assert.NoError(err, "Should decode request")

			assert.Equal(title, req.Title, "Should have correct title")
			assert.Equal(body, req.Body, "Should have correct body")
			assert.Empty(req.Head, "Should not have a head")
			assert.Equal(base, req.Base, "Should have correct base")

			w.WriteHeader(status)
			_, _ = w.Write([]byte(`{"html_url": "testurl"}`))
		}))
	}

	t.Run("Success", func(t *testing.T) {
		assert := assert.New(t)

		s := newServer(t, http.StatusOK)
		defer s.Close()

		client := NewClient(s.URL)

		res, err := client.UpdatePullRequest(token, repo, number, title, body, base)
		assert.NoError(err, "Should create PR")
		assert.Equal("testurl", res.HtmlUrl, "Should return correct PR")
	})
	t.Run("Failure", func(t *testing.T) {
		assert := assert.New(t)

		s := newServer(t, http.StatusInternalServerError)
		defer s.Close()

		client := NewClient(s.URL)

		res, err := client.UpdatePullRequest(token, repo, number, title, body, base)
		assert.Error(err, "Should fail")
		assert.Nil(res, "Should not return a PR")
	})
}

func TestCreateOrUpdatePullRequest(t *testing.T) {
	var statusCodes, expectedEndpoints []int
	var existingPR *PrResponse
	i := 0

	token, repo, title, body, branch, base := "testtoken", "test/repo", "Test Title", "Test Body", "testbranch", "testbase"

	templatePR := PrResponse{
		Number: 1234,
		Title:  title,
		Body:   body,
	}
	templatePR.Base.Ref = base
	templatePR.Head.Ref = branch

	newServer := func(t *testing.T) *httptest.Server {
		return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert := assert.New(t)

			defer func() {
				i++
			}()

			assert.NotEmpty(r.Header.Get("Accept"), "Should set Accept header")

			var req PrRequest

			switch expectedEndpoints[i] {
			case 0:
				assert.Equal(http.MethodGet, r.Method, "Should use GET method")
				assert.Equal("/repos/test/repo/pulls", r.URL.Path, "Should use correct path")

				w.WriteHeader(statusCodes[i])
				res := []PrResponse{}
				if existingPR != nil {
					res = append(res, *existingPR)
				}
				err := json.NewEncoder(w).Encode(res)
				assert.NoError(err, "Should encode response body")

				return
			case 1:
				assert.Equal(http.MethodPost, r.Method, "Should use POST method")
				assert.Equal("/repos/test/repo/pulls", r.URL.Path, "Should use correct path")

				err := json.NewDecoder(r.Body).Decode(&req)
				assert.NoError(err, "Should decode request")
			case 2:
				assert.Equal(http.MethodPatch, r.Method, "Should use PATCH method")
				assert.Equal("/repos/test/repo/pulls/1234", r.URL.Path, "Should use correct path")

				err := json.NewDecoder(r.Body).Decode(&req)
				assert.NoError(err, "Should decode request")

			default:
				t.Fatal("Unexpected endpoint")
			}

			w.WriteHeader(statusCodes[i])
			err := json.NewEncoder(w).Encode(templatePR)
			assert.NoError(err, "Should encode response body")
		}))
	}

	t.Run("GetPRFail", func(t *testing.T) {
		assert := assert.New(t)

		i = 0
		existingPR = nil
		expectedEndpoints = []int{0}
		statusCodes = []int{http.StatusInternalServerError}

		s := newServer(t)
		defer s.Close()
		client := NewClient(s.URL)

		pr, err := client.CreateOrUpdatePullRequest(token, repo, title, body, branch, base)
		assert.Nil(pr, "Should not return a PR")
		assert.ErrorContains(err, "failed to check for existing PR:", "Should fail")

		assert.Equal(1, i, "Should have made the expected api calls")
	})
	t.Run("CreatePR", func(t *testing.T) {
		assert := assert.New(t)

		i = 0
		existingPR = nil
		expectedEndpoints = []int{0, 1}
		statusCodes = []int{http.StatusOK, http.StatusCreated}

		s := newServer(t)
		defer s.Close()
		client := NewClient(s.URL)

		pr, err := client.CreateOrUpdatePullRequest(token, repo, title, body, branch, base)
		assert.NotNil(pr, "Should return a PR")
		assert.NoError(err, "Should not fail")

		assert.Equal(2, i, "Should have made the expected api calls")
	})
	t.Run("SkipUpdate", func(t *testing.T) {
		assert := assert.New(t)

		i = 0
		existingPR = &templatePR
		expectedEndpoints = []int{0}
		statusCodes = []int{http.StatusOK}

		s := newServer(t)
		defer s.Close()
		client := NewClient(s.URL)

		pr, err := client.CreateOrUpdatePullRequest(token, repo, title, body, branch, base)
		assert.NotNil(pr, "Should return a PR")
		assert.NoError(err, "Should not fail")

		assert.Equal(1, i, "Should have made the expected api calls")
	})
	t.Run("UpdatePR", func(t *testing.T) {
		assert := assert.New(t)

		i = 0
		existingPR = &templatePR
		expectedEndpoints = []int{0, 2}
		statusCodes = []int{http.StatusOK, http.StatusOK}

		s := newServer(t)
		defer s.Close()
		client := NewClient(s.URL)

		oldTitle := title
		t.Cleanup(func() {
			title = oldTitle
		})
		title = "New Title"

		pr, err := client.CreateOrUpdatePullRequest(token, repo, title, body, branch, base)
		assert.NotNil(pr, "Should return a PR")
		assert.NoError(err, "Should not fail")

		assert.Equal(2, i, "Should have made the expected api calls")
	})
}
