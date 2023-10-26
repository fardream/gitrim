// svc contains an implementation of the gRPC service
package svc

import (
	"crypto/cipher"
	"net/http"
	"sync"

	"go.etcd.io/bbolt"
)

// Svc implements the gRPC service.
//
// Svc uses [bbolt.DB] as an underlying database, and the current
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

	dbmutex sync.Mutex
	idmutex map[string]*waitingChan
}

var _ GiTrimServer = (*Svc)(nil)
