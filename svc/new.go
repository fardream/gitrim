package svc

func New(cfg *GiTrimConfig) (*Svc, error) {
	svc := &Svc{config: cfg}
	if svc.config.Remotes == nil {
		svc.config.Remotes = make(map[string]*RemoteConfig)
	}
	if err := svc.setupDb(); err != nil {
		return nil, err
	}

	if err := svc.setupCipher(); err != nil {
		return nil, err
	}

	return svc, nil
}
