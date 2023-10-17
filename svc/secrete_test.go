package svc

import (
	"crypto/aes"
	"crypto/cipher"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestSVC_newSecret(t *testing.T) {
	secret := []byte("this is a secret")
	toencode := []byte("this is a secret message")

	b, err := aes.NewCipher(secret)
	if err != nil {
		t.Fatal(err)
	}
	g, err := cipher.NewGCM(b)
	if err != nil {
		t.Fatal(err)
	}

	data, err := newSecret(g, toencode)
	if err != nil {
		t.Fatal(err)
	}

	decoded, err := decodeSecret(g, data)
	if err != nil {
		t.Fatal(err)
	}

	if !cmp.Equal(decoded, toencode) {
		t.Fatalf("want: %v, got: %v", toencode, decoded)
	}
}
