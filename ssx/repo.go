package ssx

import (
	"github.com/vimiix/ssx/ssx/bbolt"
	"github.com/vimiix/ssx/ssx/entry"
)

// Repo define a KV store interface
type Repo interface {
	Open() error
	Close() error
	GetMetadata(key []byte) ([]byte, error)
	SetMetadata(key []byte, value []byte) error
	UpdateEntry(t *entry.Entry) (err error)
	GetEntry(ip, user string) (t *entry.Entry, err error)
	GetAllEntries() (map[string]*entry.Entry, error)
	DeleteEntry(ip, user string) error
}

var _ Repo = (*bbolt.Repo)(nil)
