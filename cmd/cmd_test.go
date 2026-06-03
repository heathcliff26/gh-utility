package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewRootCommand(t *testing.T) {
	cmd := NewRoot()

	assert.Equal(t, Name, cmd.Use)
}
