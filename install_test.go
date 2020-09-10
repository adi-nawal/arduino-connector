package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInstallDockerPath(t *testing.T) {
	value, err := isDockerInstalled()
	assert.True(t, value)
	assert.True(t, err == nil)
}
