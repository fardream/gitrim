package svc

import (
	"github.com/go-git/go-git/v5/plumbing"

	"github.com/fardream/gitrim"
)

func (s *SyncStat) IsEmpty() bool {
	return s == nil || s.LastSyncFromCommit == ""
}

func (s *SyncStat) SetToEmpty() {
	s.Reset()
}

func (s *SyncStat) Hashes() (fromhead plumbing.Hash, frompastcommits gitrim.HashSet, tohead plumbing.Hash, topastcommits gitrim.HashSet, err error) {
	if s.IsEmpty() {
		return
	}

	fromhead, err = gitrim.DecodeHashHex(s.LastSyncFromCommit)
	if err != nil {
		return
	}
	frompastcommits, err = gitrim.NewHashSetFromStrings(s.FromDfs...)
	if err != nil {
		return
	}
	tohead, err = gitrim.DecodeHashHex(s.LastSyncToCommit)
	if err != nil {
		return
	}
	topastcommits, err = gitrim.NewHashSetFromStrings(s.ToDfs...)
	if err != nil {
		return
	}

	return
}

func EmptySyncStat() *SyncStat {
	return &SyncStat{
		FromToTo: make(map[string]string),
		ToToFrom: make(map[string]string),
	}
}
