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

type Svc interface {
	GiTrimServer

	Start(context.Context) error
}

type svc struct {
	// config of the server.
	config *GiTrimConfig

	// db of the process
	db        *bbolt.DB
	tmpDbPath string

	// listener to webhooks.
	webhookMutex *http.ServeMux

	// we are going to risk it.
	UnsafeGiTrimServer

	idBlock cipher.Block
}

func (*svc) SyncToSubRepo(context.Context, *SyncToSubRepoRequest) (*SyncToSubRepoResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SyncToSubRepo not implemented")
}

func (*svc) CommitFromSubRepo(context.Context, *CommitFromSubRepoRequest) (*CommitFromSubRepoResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CommitFromSubRepo not implemented")
}

func (*svc) CheckCommitFromSubRepo(context.Context, *CheckCommitFromSubRepoRequest) (*CheckCommitFromSubRepoResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CheckCommitFromSubRepo not implemented")
}

var _ GiTrimServer = (*svc)(nil)

func New(cfg *GiTrimConfig) (Svc, error) {
	svc := &svc{config: cfg}
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

func (s *svc) HttpServerMux() (*http.ServeMux, error) {
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

func (s *svc) cleanup() error {
	if err := s.cleanUpDb(); err != nil {
		return err
	}
	return nil
}

// Start a server
func (s *svc) Start(ctx context.Context) error {
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
		s.cleanup()
		return err
	}

	wg.Wait()

	s.cleanup()

	return nil
}

func (s *svc) HandleNewPullRequest(ctx context.Context, pr *PullRequestInfo) error {
	return nil
}

func (s *svc) DeleteTmpDb() error {
	s.cleanup()
	if s.tmpDbPath == "" {
		return nil
	}
	return os.RemoveAll(s.tmpDbPath)
}
