package svc

func (s *Svc) Close() error {
	if err := s.closeDb(); err != nil {
		return err
	}
	return nil
}
