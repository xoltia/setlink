package models

import (
	"crypto/sha256"
	"encoding/hex"
	"net/url"

	"gorm.io/gorm"
)

type Link struct {
	ID          uint       `json:"id" gorm:"primarykey"`
	URL         string     `json:"url"`
	Favicon     string     `json:"favicon"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Image       string     `json:"image"`
	LinkSets    []*LinkSet `json:"-" gorm:"many2many:link_set_links;"`
	HashString  string     `json:"hash" gorm:"uniqueIndex"`
}

func (l *Link) BeforeCreate(db *gorm.DB) (err error) {
	l.ComputeHash()
	return
}

func (l *Link) ComputeHash() ([32]byte, error) {
	if l.HashString != "" {
		hashBytes, err := hex.DecodeString(l.HashString)
		if err != nil {
			return [32]byte{}, err
		}

		var hash32 [32]byte
		copy(hash32[:], hashBytes)
		return hash32, nil
	}

	url, err := url.Parse(l.URL)

	if err != nil {
		return [32]byte{}, err
	}

	hash := sha256.New()

	hash.Write([]byte(url.Host))
	hash.Write([]byte(url.Path))
	// Query string should be ordered
	hash.Write([]byte(url.Query().Encode()))

	var hash32 [32]byte
	copy(hash32[:], hash.Sum(nil))

	l.HashString = hex.EncodeToString(hash32[:])

	return hash32, nil
}
