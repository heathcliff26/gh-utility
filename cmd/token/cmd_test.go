package token

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRequiredFlags(t *testing.T) {
	assert := assert.New(t)

	cmd := NewCommand()

	err := cmd.ValidateRequiredFlags()
	assert.ErrorContains(err, "required flag(s)")
	assert.ErrorContains(err, appKeyPathFlag)
	assert.ErrorContains(err, clientIDFlag)
	assert.ErrorContains(err, installationIDFlag)
}

func TestRun(t *testing.T) {
	assert := assert.New(t)

	cmd := NewCommand()

	err := run(cmd)
	assert.Error(err)
}
