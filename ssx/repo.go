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
	GetEntry(id uint64) (*entry.Entry, error)
	GetAllEntries() (map[uint64]*entry.Entry, error)
	DeleteEntry(id uint64) error
}

var _ Repo = (*bbolt.Repo)(nil)
