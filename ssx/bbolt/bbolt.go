package bbolt

import (
	"encoding/json"
	"fmt"

	"go.etcd.io/bbolt"

	"github.com/vimiix/ssx/internal/errmsg"
	"github.com/vimiix/ssx/internal/lg"
	"github.com/vimiix/ssx/ssx/entry"
)

// entryKey returns formatted unique key of Entry
func entryKey(ip, user string) []byte {
	s := fmt.Sprintf("%s/%s", ip, user)
	return []byte(s)
}

type Repo struct {
	db          *bbolt.DB
	file        string
	metaBucket  []byte
	entryBucket []byte
}

func (r *Repo) GetMetadata(key []byte) ([]byte, error) {
	if r.db == nil {
		return nil, errmsg.ErrRepoNotOpen
	}
	var res []byte
	_ = r.db.View(func(tx *bbolt.Tx) error {
		res = tx.Bucket(r.metaBucket).Get(key)
		return nil
	})
	return res, nil
}

func (r *Repo) SetMetadata(key []byte, value []byte) error {
	if r.db == nil {
		return errmsg.ErrRepoNotOpen
	}
	return r.db.Update(func(tx *bbolt.Tx) error {
		return tx.Bucket(r.metaBucket).Put(key, value)
	})
}

func (r *Repo) UpdateEntry(e *entry.Entry) (err error) {
	if r.db == nil {
		err = errmsg.ErrRepoNotOpen
		return
	}

	return r.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(r.entryBucket)
		key := entryKey(e.Host, e.User)
		bs := b.Get(key)
		if bs == nil {
			// insert
			e.ID, _ = b.NextSequence()
		}
		// update
		buf, marshalErr := json.Marshal(e)
		if marshalErr != nil {
			return marshalErr
		}
		return b.Put(key, buf)
	})
}

func (r *Repo) GetEntry(ip, user string) (t *entry.Entry, err error) {
	if r.db == nil {
		err = errmsg.ErrRepoNotOpen
		return
	}

	err = r.db.View(func(tx *bbolt.Tx) error {
		bs := tx.Bucket(r.entryBucket).Get(entryKey(ip, user))
		if bs == nil {
			return errmsg.ErrEntryNotExist
		}
		return json.Unmarshal(bs, t)
	})
	return
}

// GetAllEntries returns all entries map, key format is "ip/user"
func (r *Repo) GetAllEntries() (map[string]*entry.Entry, error) {
	if r.db == nil {
		return nil, errmsg.ErrRepoNotOpen
	}

	var (
		err error
		m   = map[string]*entry.Entry{}
	)

	err = r.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(r.entryBucket)
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			var t = entry.Entry{}
			if unmarshalErr := json.Unmarshal(v, &t); unmarshalErr != nil {
				return unmarshalErr
			}
			m[string(k)] = &t
		}
		return nil
	})
	return m, err
}

func (r *Repo) DeleteEntry(ip, user string) error {
	if r.db == nil {
		return errmsg.ErrRepoNotOpen
	}

	return r.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(r.entryBucket)
		return b.Delete(entryKey(ip, user))
	})
}

func (r *Repo) Open() error {
	db, err := bbolt.Open(r.file, 0600, nil)
	if err != nil {
		return err
	}
	err = db.Update(func(tx *bbolt.Tx) error {
		for _, bucketName := range r.buckets() {
			_, createErr := tx.CreateBucketIfNotExists(bucketName)
			if createErr != nil {
				return createErr
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	r.db = db
	return nil
}

func (r *Repo) Close() error {
	if r.db == nil {
		return nil
	}
	err := r.db.Close()
	if err == nil {
		r.db = nil
	}
	return err
}

func (r *Repo) buckets() [][]byte {
	return [][]byte{r.metaBucket, r.entryBucket}
}

func NewRepo(file string) *Repo {
	lg.Debug("new repo with %q", file)
	return &Repo{
		file:        file,
		metaBucket:  []byte("metadata"),
		entryBucket: []byte("entries"),
	}
}
