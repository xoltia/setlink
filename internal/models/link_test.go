package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLinkHash(t *testing.T) {
	link := Link{
		URL: "https://example.com/path?query=string&another=one",
	}

	hash, err := link.ComputeHash()

	assert.Nil(t, err)

	similarLink := Link{
		URL: "https://example.com/path?another=one&query=string",
	}

	sameHash, err := similarLink.ComputeHash()

	assert.Nil(t, err)
	assert.Equal(t, hash, sameHash)
}
