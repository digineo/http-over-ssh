package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKeyFile(t *testing.T) {
	t.Parallel()

	_, err := getKeyFile("fixtures/id_ed25519")
	assert.NoError(t, err)

	_, err = getKeyFile("does-not-exist")
	assert.EqualError(t, err, "open does-not-exist: no such file or directory")
}

func TestReadPrivateKeys(t *testing.T) {
	t.Parallel()

	keys := readPrivateKeys("does-not-exists", "fixtures/id_ed25519")
	assert.Len(t, keys, 1)
}
