// svc contains a service that can handle filtering repos on different git providers.
package svc

import (
	"context"
	"crypto/cipher"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	"go.etcd.io/bbolt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Svc struct {
	// config of the server.
	config *GiTrimConfig

	// db of the process
	db        *bbolt.DB
	tmpDbPath string

	// listener to webhooks.
	webhookMutex *http.ServeMux

	// we are going to risk it.
	UnsafeGiTrimServer

	encryptor cipher.AEAD
}

func (*Svc) SyncToSubRepo(context.Context, *SyncToSubRepoRequest) (*SyncToSubRepoResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SyncToSubRepo not implemented")
}

func (*Svc) CommitFromSubRepo(context.Context, *CommitFromSubRepoRequest) (*CommitFromSubRepoResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CommitFromSubRepo not implemented")
}

func (*Svc) CheckCommitFromSubRepo(context.Context, *CheckCommitFromSubRepoRequest) (*CheckCommitFromSubRepoResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CheckCommitFromSubRepo not implemented")
}

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

func (s *Svc) HttpServerMux() (*http.ServeMux, error) {
	if s.webhookMutex != nil {
		return s.webhookMutex, nil
	}

	m := http.NewServeMux()
	s.webhookMutex = m

	m.HandleFunc("/pr-hook", func(w http.ResponseWriter, r *http.Request) {
		bs, err := io.ReadAll(r.Body)
		if err != nil {
			slog.Error("failed to read body", "err", err.Error())
		}

		os.Stdout.Write(bs)

		fmt.Fprint(w, "we got it!")
	})

	return m, nil
}

func (s *Svc) Close() error {
	if err := s.closeDb(); err != nil {
		return err
	}
	return nil
}

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

		slog.Info("shutdown requested")

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

	slog.Info("launching webhooks", "addr", webhookSvc.Addr)

	if err := webhookSvc.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		s.Close()
		return err
	}

	wg.Wait()

	s.Close()

	return nil
}

func (s *Svc) HandleNewPullRequest(ctx context.Context, pr *PullRequestInfo) error {
	return nil
}
