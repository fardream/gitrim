package svc

import (
	"crypto/cipher"
	"crypto/rand"
	"fmt"
)

// newSecret creates a new secret for a given id
func newSecret(encrypto cipher.AEAD, id []byte) ([]byte, error) {
	nonce := make([]byte, encrypto.NonceSize())
	n, err := rand.Read(nonce)
	if err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}
	if n != encrypto.NonceSize() {
		return nil, fmt.Errorf("failed to generate enough nonce size")
	}

	return encrypto.Seal(nonce, nonce, id, nil), nil
}

// decodeSecret gets the
func decodeSecret(encryptor cipher.AEAD, secret []byte) ([]byte, error) {
	if len(secret) <= encryptor.NonceSize() {
		return nil, fmt.Errorf("the length of secret is too small")
	}

	nonce, encrypted := secret[:encryptor.NonceSize()], secret[encryptor.NonceSize():]

	id, err := encryptor.Open(nil, nonce, encrypted, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to unseal id: %w", err)
	}

	return id, nil
}
