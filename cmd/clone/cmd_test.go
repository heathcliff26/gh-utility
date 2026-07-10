package clone

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewCommand(t *testing.T) {
	assert := assert.New(t)

	cmd := NewCommand()

	err := cmd.ValidateRequiredFlags()
	assert.NoError(err, "Should not have any required flags")

	err = cmd.ValidateArgs([]string{})
	assert.Error(err, "Should require at least 1 argument")

	err = cmd.ValidateArgs([]string{"repo"})
	assert.NoError(err, "Should accept 1 argument")

	err = cmd.ValidateArgs([]string{"repo", "dir"})
	assert.NoError(err, "Should accept 2 arguments")

	err = cmd.ValidateArgs([]string{"repo", "dir", "extra"})
	assert.Error(err, "Should not accept more than 2 arguments")
}

func TestRun(t *testing.T) {
	assert := assert.New(t)

	cmd := NewCommand()

	flags := []string{fetchDepthFlag, "1"}
	_ = cmd.ParseFlags(flags)

	tmpDir := t.TempDir()

	args := []string{"https://github.com/heathcliff26/gh-utility.git", tmpDir}

	err := run(cmd, args)
	assert.NoError(err, "Should run without error")

	ls, err := os.ReadDir(tmpDir)
	assert.NoError(err, "Should read directory")
	assert.GreaterOrEqual(len(ls), 1, "Should have at least one entry")
}
