package svc

import (
	"context"
	"errors"
	"net"
	"net/http"
	sync "sync"
	"time"
)

// Start a server
func (s *Svc) Start(ctx context.Context) error {
	webhookHandler, err := s.HttpServerMux()
	if err != nil {
		return err
	}
	var wg sync.WaitGroup

	webhookSvc := &http.Server{
		Addr:    s.config.WebhookAddress,
		Handler: webhookHandler,
		BaseContext: func(net.Listener) context.Context {
			return ctx
		},
	}

	wg.Add(1)
	go func() {
		<-ctx.Done()

		logger.Info("shutdown requested")

		ctx, cancel := context.WithTimeout(
			context.Background(),
			time.Second*time.Duration(
				s.config.GetProperShutdownWaitSecs(),
			),
		)
		defer cancel()
		webhookSvc.Shutdown(ctx)

		wg.Done()
	}()

	logger.Info("launching webhooks", "addr", webhookSvc.Addr)

	if err := webhookSvc.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		s.Close()
		return err
	}

	wg.Wait()

	s.Close()

	return nil
}
