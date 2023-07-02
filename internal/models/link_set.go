package models

import (
	"encoding/hex"

	"gorm.io/gorm"
)

type LinkSet struct {
	ID         uint    `json:"id" gorm:"primarykey"`
	HashString string  `json:"hash" gorm:"uniqueIndex"`
	Links      []*Link `json:"links" gorm:"many2many:link_set_links;"`
}

func (l *LinkSet) BeforeCreate(db *gorm.DB) (err error) {
	l.ComputeHash()
	return
}

func (l *LinkSet) ComputeHash() ([32]byte, error) {
	if l.HashString != "" {
		hashBytes, err := hex.DecodeString(l.HashString)
		if err != nil {
			return [32]byte{}, err
		}

		var hash32 [32]byte
		copy(hash32[:], hashBytes)
		return hash32, nil
	}

	var hash [32]byte

	for _, link := range l.Links {
		linkHash, err := link.ComputeHash()

		if err != nil {
			return [32]byte{}, err
		}

		if hash == [32]byte{} {
			hash = linkHash
		} else {
			for i := 0; i < 32; i++ {
				hash[i] ^= linkHash[i]
			}
		}
	}

	l.HashString = hex.EncodeToString(hash[:])

	return hash, nil
}
