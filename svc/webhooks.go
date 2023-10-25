package svc

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

func (s *Svc) HttpServerMux() (*http.ServeMux, error) {
	if s.webhookMutex != nil {
		return s.webhookMutex, nil
	}

	m := http.NewServeMux()
	s.webhookMutex = m

	m.HandleFunc("/pr-hook", func(w http.ResponseWriter, r *http.Request) {
		bs, err := io.ReadAll(r.Body)
		if err != nil {
			logger.Error("failed to read body", "err", err.Error())
		}

		os.Stdout.Write(bs)

		fmt.Fprint(w, "we got it!")
	})

	return m, nil
}
