package svc

import (
	"errors"
	"fmt"
)

func (s *Svc) FindRemoteInfo(name string) (*RemoteConfig, error) {
	if s.config == nil || s.config.Remotes == nil {
		return nil, errors.New("empty remotes")
	}

	remote, find := s.config.Remotes[name]

	if !find {
		return nil, fmt.Errorf("cannot find remote named %s", name)
	}

	return remote, nil
}
