package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInstallDockerPath(t *testing.T) {
	str, err := createDockerConfig()
	assert.True(t, str == "installed")
	assert.True(t, err == nil)
}
