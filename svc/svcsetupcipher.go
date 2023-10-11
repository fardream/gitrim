package svc

import (
	"crypto/aes"
	"encoding/hex"
	"fmt"
	"log/slog"
)

var zeroKey []byte = make([]byte, 16)

func (s *svc) setupCipher() error {
	keyHex := s.config.AesKey
	if keyHex == "" {
		slog.Warn("empty cipher key")

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

	s.idBlock = b

	return nil
}
