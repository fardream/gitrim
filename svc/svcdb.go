package svc

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path"

	"go.etcd.io/bbolt"
	"google.golang.org/protobuf/proto"
)

func (s *svc) setupDb() error {
	dbpath := s.config.DbPath
	var err error
	if dbpath == "" {
		dbpathfolder, err := os.MkdirTemp("", "gitrim-*")
		if err != nil {
			return fmt.Errorf("config doesn't have db path but failed to create a tmp: %w", err)
		}
		s.tmpDbPath = path.Join(dbpathfolder, "gitrim.db")
		dbpath = s.tmpDbPath
		slog.Warn("missing db path, use tmp path", "path", dbpath)
	}

	db, err := bbolt.Open(dbpath, 0o600, nil)
	if err != nil {
		return err
	}

	s.db = db

	return nil
}

func (s *svc) cleanUpDb() error {
	if s.db == nil {
		return nil
	}

	if s.tmpDbPath != "" {
		slog.Warn("missing db path, used tmp path", "path", s.tmpDbPath)
	}

	return s.db.Close()
}

func (s *svc) mustGetDb() *bbolt.DB {
	if s.db == nil {
		slog.Error("no db")
		panic(errors.New("no db setup"))
	}

	return s.db
}

const RepoSyncBucket = "reposyncs"

func getRepoSyncFromDb(db *bbolt.DB, id []byte) (*RepoSync, error) {
	var d []byte = nil

	if err := db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(RepoSyncBucket))
		if b == nil {
			return nil
		}
		v := b.Get(id)
		if v != nil {
			d = v
		}

		return nil
	}); err != nil {
		return nil, err
	}

	if d == nil {
		return nil, nil
	}

	r := &RepoSync{}
	if err := proto.Unmarshal(d, r); err != nil {
		return nil, err
	}

	return r, nil
}
