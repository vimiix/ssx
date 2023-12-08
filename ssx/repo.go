package ssx

import (
	"github.com/vimiix/ssx/ssx/bbolt"
	"github.com/vimiix/ssx/ssx/entry"
)

// Repo define a KV store interface
type Repo interface {
	Init() error
	GetMetadata(key []byte) ([]byte, error)
	SetMetadata(key []byte, value []byte) error
	TouchEntry(e *entry.Entry) (err error)
	GetEntry(ip, user string) (e *entry.Entry, err error)
	GetAllEntries() (map[string]*entry.Entry, error)
	DeleteEntry(ip, user string) error
}

var _ Repo = (*bbolt.Repo)(nil)
