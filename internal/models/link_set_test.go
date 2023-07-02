package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestLinkSetHash(t *testing.T) {
	linkSet := LinkSet{
		Links: []*Link{
			{
				URL: "https://example.com/path?query=string&another=one",
			},
			{
				URL: "https://example.com/path?another=one&query=string",
			},
		},
	}

	hash, err := linkSet.ComputeHash()

	assert.Nil(t, err)

	similarLinkSet := LinkSet{
		Links: []*Link{
			{
				URL: "https://example.com/path?another=one&query=string",
			},
			{
				URL: "https://example.com/path?query=string&another=one",
			},
		},
	}

	sameHash, err := similarLinkSet.ComputeHash()

	assert.Nil(t, err)
	assert.Equal(t, hash, sameHash)
}

func TestLinkSetCreate(t *testing.T) {
	// create in memory sqlite database
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	linkSet := LinkSet{
		Links: []*Link{
			{
				URL: "https://example.com/path?query=string&another=one",
			},
			{
				URL: "https://another-example.com/path2?another=one&query=string",
			},
		},
	}

	db.AutoMigrate(&LinkSet{}, &Link{})

	db.Create(&linkSet)
	db.First(&linkSet)

	assert.NotEmpty(t, linkSet.HashString)
	assert.NotEmpty(t, linkSet.ID)
}
