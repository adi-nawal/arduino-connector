package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInstallUbuntuDockerPath(t *testing.T) {
	value, err := isDockerInstalled()
	assert.False(t, value)
	assert.True(t, err == nil)
}

func TestInstallAlpineDockerPath(t *testing.T) {
	value, err := isDockerInstalled()
	assert.True(t, value)
	assert.True(t, err == nil)
}

func TestInstallUbuntuNetworkManager(t *testing.T) {
	assert.False(t, isNetManagerInstalled())
}
