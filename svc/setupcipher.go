package svc

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"fmt"
)

var zeroKey []byte = make([]byte, 16)

// setupCipher
func (s *Svc) setupCipher() error {
	keyHex := s.config.AesKey
	if keyHex == "" {
		logger.Warn("empty cipher key")
		keyHex = hex.EncodeToString(zeroKey)
	}

	key, err := hex.DecodeString(keyHex)
	if err != nil {
		return fmt.Errorf("failed to parse hex for key: %w", err)
	}

	if len(key) != aes.BlockSize {
		return fmt.Errorf("length of parse key %d is not right", len(key))
	}

	b, err := aes.NewCipher(key)
	if err != nil {
		return fmt.Errorf("failed to create id block: %w", err)
	}

	v, err := cipher.NewGCM(b)
	if err != nil {
		return fmt.Errorf("failed to create GCM: %w", err)
	}

	s.encryptor = v

	return nil
}
