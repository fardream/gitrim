package svc

import (
	"log/slog"
	"os"

	"go.etcd.io/bbolt"
	"google.golang.org/protobuf/proto"
)

// getFromDb returns the typed
func getFromDb[
	T any](db *bbolt.DB, bucket []byte, id []byte,
	unmarshal func(data []byte, v *T) error,
) (*T, error) {
	if db == nil {
		return nil, ErrNilDB
	}

	r := (*T)(nil)

	err := db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(bucket)
		if b == nil {
			return nil
		}
		v := b.Get(id)
		if v == nil {
			return nil
		}
		r = new(T)
		if err := unmarshal(v, r); err != nil {
			r = nil
			return err
		}

		return nil
	})

	return r, err
}

func putToDb[T any](db *bbolt.DB, bucket []byte, id []byte, v T, marshal func(v T) ([]byte, error)) error {
	if db == nil {
		return ErrNilDB
	}

	return db.Update(
		func(tx *bbolt.Tx) error {
			b, err := tx.CreateBucketIfNotExists(bucket)
			if err != nil {
				return err
			}
			data, err := marshal(v)
			if err != nil {
				return err
			}
			return b.Put(id, data)
		})
}

// tempfile provides a temporary file, adopted from the example on [bbolt doc]
//
// [bbolt doc]: https://pkg.go.dev/go.etcd.io/bbolt#example-DB.Begin
func tempfile() (string, error) {
	f, err := os.CreateTemp("", "bolt-")
	if err != nil {
		return "", err
	}
	if err := f.Close(); err != nil {
		return "", err
	}
	if err := os.Remove(f.Name()); err != nil {
		return "", err
	}
	return f.Name(), nil
}

func (s *Svc) setupDb() error {
	dbpath := s.config.DbPath
	var err error
	if dbpath == "" {
		dbpath, err = tempfile()
		s.tmpDbPath = dbpath
		slog.Warn("missing db path, use tmp path", "path", dbpath)
	}

	db, err := bbolt.Open(dbpath, 0o600, nil)
	if err != nil {
		return err
	}

	s.db = db

	return nil
}

func (s *Svc) closeDb() error {
	if s.db == nil {
		return nil
	}

	if s.tmpDbPath != "" {
		slog.Warn("missing db path, used tmp path", "path", s.tmpDbPath)
	}

	return s.db.Close()
}

// mustGetDb returns the database for the service, panics if the database is nil.
func (s *Svc) mustGetDb() *bbolt.DB {
	if s.db == nil {
		slog.Error("no db")
		panic(ErrNilDB)
	}

	return s.db
}

func (s *Svc) DeleteTmpDb() error {
	s.Close()
	if s.tmpDbPath == "" {
		return nil
	}
	return os.Remove(s.tmpDbPath)
}

const (
	REPO_SYNC_BUCKET    = "reposyncs"
	SECRET_TO_ID_BUCKET = "secrets-to-id"
	ID_TO_SECRET_BUCKET = "id-to-secrets"
)

func getRepoSyncFromDb(
	db *bbolt.DB,
	id []byte,
) (*RepoSync, error) {
	return getFromDb(
		db,
		[]byte(REPO_SYNC_BUCKET),
		id,
		func(d []byte, v *RepoSync) error {
			return proto.Unmarshal(d, v)
		})
}

func putRepoSyncToDb(
	db *bbolt.DB,
	id []byte,
	r *RepoSync,
) error {
	return putToDb(
		db,
		[]byte(REPO_SYNC_BUCKET),
		id,
		r,
		func(v *RepoSync) ([]byte, error) {
			return proto.Marshal(v)
		})
}

func passByte(i []byte) ([]byte, error) {
	return i, nil
}

func putSecretFunc(id []byte, secret []byte) func(tx *bbolt.Tx) error {
	return func(tx *bbolt.Tx) error {
		idtosecretbucket, err := tx.CreateBucketIfNotExists([]byte(ID_TO_SECRET_BUCKET))
		if err != nil {
			return err
		}
		if err := idtosecretbucket.Put(id, secret); err != nil {
			return err
		}
		secrettoidbucket, err := tx.CreateBucketIfNotExists([]byte(SECRET_TO_ID_BUCKET))
		if err != nil {
			return err
		}
		return secrettoidbucket.Put(secret, id)
	}
}

func putRepoSyncFunc(id []byte, reposync *RepoSync) func(tx *bbolt.Tx) error {
	return func(tx *bbolt.Tx) error {
		reposyncbucket, err := tx.CreateBucketIfNotExists([]byte(REPO_SYNC_BUCKET))
		if err != nil {
			return err
		}

		b, err := proto.Marshal(reposync)
		if err != nil {
			return err
		}

		return reposyncbucket.Put(id, b)
	}
}
