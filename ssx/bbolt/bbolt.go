package bbolt

import (
	"encoding/binary"
	"encoding/json"
	"time"

	"go.etcd.io/bbolt"

	"github.com/vimiix/ssx/internal/encrypt"
	"github.com/vimiix/ssx/internal/errmsg"
	"github.com/vimiix/ssx/internal/lg"
	"github.com/vimiix/ssx/ssx/entry"
)

// itob returns an 8-byte big endian representation of v.
func itob(v uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, v)
	return b
}

type Repo struct {
	db          *bbolt.DB
	file        string
	metaBucket  []byte
	entryBucket []byte
}

func (r *Repo) GetMetadata(key []byte) ([]byte, error) {
	if err := r.open(); err != nil {
		return nil, err
	}
	defer r.close()
	var res []byte
	lg.Debug("bbolt repo: get metadata: %s", string(key))
	_ = r.db.View(func(tx *bbolt.Tx) error {
		res = tx.Bucket(r.metaBucket).Get(key)
		return nil
	})
	return res, nil
}

func (r *Repo) SetMetadata(key []byte, value []byte) error {
	if err := r.open(); err != nil {
		return err
	}
	defer r.close()
	lg.Debug("bbolt repo: set metadata: %s", string(key))
	return r.db.Update(func(tx *bbolt.Tx) error {
		return tx.Bucket(r.metaBucket).Put(key, value)
	})
}

func (r *Repo) TouchEntry(e *entry.Entry) error {
	if err := r.open(); err != nil {
		return err
	}
	defer r.close()

	lg.Debug("bbolt repo: touch entry: %d", e.ID)
	return r.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(r.entryBucket)
		var bs []byte
		if e.ID > 0 {
			bs = b.Get(itob(e.ID))
		}
		if len(bs) == 0 {
			// insert
			e.ID, _ = b.NextSequence()
			now := time.Now()
			e.VisitCount = 1
			e.CreateAt = now
			e.UpdateAt = now
		} else {
			var rawEntry = &entry.Entry{}
			if err := json.Unmarshal(bs, rawEntry); err != nil {
				return err
			}
			e.ID = rawEntry.ID
			e.VisitCount = rawEntry.VisitCount + 1
			e.CreateAt = rawEntry.CreateAt
			e.UpdateAt = time.Now()
		}
		// update
		buf, encodeErr := encodeEntry(e)
		if encodeErr != nil {
			return encodeErr
		}
		return b.Put(itob(e.ID), buf)
	})
}

func (r *Repo) GetEntry(id uint64) (e *entry.Entry, err error) {
	if err = r.open(); err != nil {
		return
	}
	defer r.close()

	lg.Debug("bbolt repo: get entry by id: %d", id)
	err = r.db.View(func(tx *bbolt.Tx) error {
		bs := tx.Bucket(r.entryBucket).Get(itob(id))
		if len(bs) == 0 {
			return errmsg.ErrEntryNotExist
		}
		var decodeErr error
		e, decodeErr = decodeEntry(bs)
		if decodeErr != nil {
			return decodeErr
		}
		return nil
	})
	return
}

// GetAllEntries returns all entries map, key format is "ip/user"
func (r *Repo) GetAllEntries() (map[uint64]*entry.Entry, error) {
	if err := r.open(); err != nil {
		return nil, err
	}
	defer r.close()

	var (
		err error
		m   = map[uint64]*entry.Entry{}
	)

	lg.Debug("bbolt repo: get all enrties")
	err = r.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(r.entryBucket)
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			e, decodeErr := decodeEntry(v)
			if decodeErr != nil {
				return decodeErr
			}
			m[e.ID] = e
		}
		return nil
	})
	return m, err
}

func (r *Repo) DeleteEntry(id uint64) error {
	if err := r.open(); err != nil {
		return err
	}
	defer r.close()

	lg.Debug("bbolt repo: delete entry: %d", id)
	return r.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(r.entryBucket)
		return b.Delete(itob(id))
	})
}

func (r *Repo) Init() error {
	if err := r.open(); err != nil {
		return err
	}
	defer r.close()
	return r.db.Update(func(tx *bbolt.Tx) error {
		for _, bucketName := range r.buckets() {
			_, createErr := tx.CreateBucketIfNotExists(bucketName)
			if createErr != nil {
				return createErr
			}
		}
		return nil
	})
}

func (r *Repo) close() error {
	if r.db == nil {
		return nil
	}
	err := r.db.Close()
	if err == nil {
		r.db = nil
	}
	return err
}

func (r *Repo) open() error {
	db, err := bbolt.Open(r.file, 0600, nil)
	if err != nil {
		return err
	}
	r.db = db
	return nil
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

func encodeEntry(e *entry.Entry) ([]byte, error) {
	e.Password = encrypt.Encrypt(e.Password)
	e.Passphrase = encrypt.Encrypt(e.Passphrase)
	return json.Marshal(e)
}

func decodeEntry(bs []byte) (*entry.Entry, error) {
	var e = &entry.Entry{}
	if err := json.Unmarshal(bs, e); err != nil {
		return nil, err
	}
	e.Password = encrypt.Decrypt(e.Password)
	e.Passphrase = encrypt.Decrypt(e.Passphrase)
	return e, nil
}
