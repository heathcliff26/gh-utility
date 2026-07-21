package client

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRequest(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	req, err := newRequest(http.MethodGet, "localhost", nil, "")
	require.NoError(err, "Should return request without error")
	require.NotNil(req, "Should return request")
	assert.Equal("application/vnd.github+json", req.Header.Get("Accept"), "Should set Accept header")
	assert.Equal("2026-03-10", req.Header.Get("X-GitHub-Api-Version"), "Should set X-GitHub-Api-Version header")
	assert.Empty(req.Header.Get("Authorization"), "Should not set Authorization header")

	req, err = newRequest(http.MethodGet, "localhost", nil, "token")
	require.NoError(err, "Should return request without error")
	require.NotNil(req, "Should return request")
	assert.Equal("Bearer token", req.Header.Get("Authorization"), "Should set Authorization")
}

func TestFileToTreeObject(t *testing.T) {
	t.Run("File", func(t *testing.T) {
		assert := assert.New(t)
		require := require.New(t)

		obj, err := fileToTreeObject("", "testdata/tree/file.txt")
		require.NoError(err)
		require.NotNil(obj)

		assert.Equal("testdata/tree/file.txt", obj.Path, "Should have correct Path")
		assert.Equal(treeTypeBlob, obj.Type, "Should have correct Type")
		assert.NotEmpty(obj.Content, "Should have read file content")
		assert.Equal(treeModeFile, obj.Mode, "Should have correct mode")
		assert.False(obj.DeleteFile, "Should not mark file for deletion")
	})
	t.Run("Executable", func(t *testing.T) {
		assert := assert.New(t)
		require := require.New(t)

		obj, err := fileToTreeObject("testdata", "tree/executable.txt")
		require.NoError(err)
		require.NotNil(obj)

		assert.Equal("tree/executable.txt", obj.Path, "Should have correct Path")
		assert.Equal(treeTypeBlob, obj.Type, "Should have correct Type")
		assert.NotEmpty(obj.Content, "Should have read file content")
		assert.Equal(treeModeExec, obj.Mode, "Should have correct mode")
		assert.False(obj.DeleteFile, "Should not mark file for deletion")
	})
	t.Run("Deleted", func(t *testing.T) {
		assert := assert.New(t)
		require := require.New(t)

		obj, err := fileToTreeObject("testdata/tree", "deleted.txt")
		require.NoError(err)
		require.NotNil(obj)

		assert.Equal("deleted.txt", obj.Path, "Should have correct Path")
		assert.Equal(treeTypeBlob, obj.Type, "Should have correct Type")
		assert.Empty(obj.Content, "Should have no content")
		assert.Equal(treeModeFile, obj.Mode, "Should have correct mode")
		assert.True(obj.DeleteFile, "Should mark file for deletion")
	})
}

func TestTranslateCheckRunStatus(t *testing.T) {
	tMatrix := []struct {
		Name        string
		Status      string
		Expected    string
		Conclusion  string
		ExpectError bool
	}{
		{"Succeeded", "succeeded", "completed", "success", false},
		{"Success", "success", "completed", "success", false},
		{"Failed", "failed", "completed", "failure", false},
		{"Failure", "failure", "completed", "failure", false},
		{"Skipped", "skipped", "completed", "skipped", false},
		{"TektonSkipped", "none", "completed", "skipped", false},
		{"Cancelled", "cancelled", "completed", "cancelled", false},
		{"Neutral", "neutral", "completed", "neutral", false},
		{"TimedOut", "timed_out", "completed", "timed_out", false},
		{"Timeout", "timeout", "completed", "timed_out", false},
		{"ActionRequired", "action_required", "completed", "action_required", false},
		{"Stale", "stale", "completed", "stale", false},
		{"Running", "running", "in_progress", "", false},
		{"InProgress", "in_progress", "in_progress", "", false},
		{"Pending", "pending", "queued", "", false},
		{"Queued", "queued", "queued", "", false},
		{"UnknownStatus", "unknown_status", "", "", true},
	}

	for _, tCase := range tMatrix {
		t.Run(tCase.Name, func(t *testing.T) {
			assert := assert.New(t)

			status, conclusion, err := translateCheckRunStatus(tCase.Status)
			if tCase.ExpectError {
				assert.Error(err, "Should return error for unknown status")
				assert.Contains(err.Error(), "unknown status", "Should contain expected error text")
			} else {
				assert.NoError(err, "Should not return error")
				assert.Equal(tCase.Expected, status, "Should have correct status")
				assert.Equal(tCase.Conclusion, conclusion, "Should have correct conclusion")
			}
		})
	}
}
