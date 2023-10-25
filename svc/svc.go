// svc contains a service that can handle filtering repos on different git providers.
package svc

import (
	"crypto/cipher"
	"net/http"

	"go.etcd.io/bbolt"
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

var _ GiTrimServer = (*Svc)(nil)
