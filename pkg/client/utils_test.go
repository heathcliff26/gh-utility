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
