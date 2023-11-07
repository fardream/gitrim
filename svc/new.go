package svc

func New(cfg *GiTrimConfig) (*Svc, error) {
	svc := &Svc{
		config:  cfg,
		idmutex: make(chan map[string]*waitingChan, 1),
	}

	svc.idmutex <- make(map[string]*waitingChan)

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
