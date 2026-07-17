package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetToken(t *testing.T) {
	t.Run("GITHUB", func(t *testing.T) {
		assert := assert.New(t)

		t.Setenv("GITHUB_TOKEN", "testtoken")

		assert.Equal("testtoken", GetToken())
	})
	t.Run("GH", func(t *testing.T) {
		assert := assert.New(t)

		t.Setenv("GH_TOKEN", "testtoken")

		assert.Equal("testtoken", GetToken())
	})
	t.Run("Both", func(t *testing.T) {
		assert := assert.New(t)

		t.Setenv("GITHUB_TOKEN", "testtoken")
		t.Setenv("GH_TOKEN", "wrong-token")

		assert.Equal("testtoken", GetToken())
	})
	t.Run("Empty", func(t *testing.T) {
		assert := assert.New(t)

		assert.Empty(GetToken())
	})
}
