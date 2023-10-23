package gitrim

import (
	"encoding/hex"
	"errors"

	"github.com/go-git/go-git/v5/plumbing"
)

var ErrHexStringTooShort = errors.New("hex encoded byte slice is too short for hash")

// DecodeHashHex decodes a hex encoded sha1 ([plumbing.Hash]).
// It differs from [plumbing.NewHash] for [plumbing.NewHash] doesn't
// check [hex.DecodeString] has error or the length of the decoded bytes.
func DecodeHashHex(str string) (plumbing.Hash, error) {
	v, err := hex.DecodeString(str)
	if err != nil {
		return plumbing.ZeroHash, err
	}
	if len(v) < 20 {
		return plumbing.ZeroHash, ErrHexStringTooShort
	}

	r := plumbing.Hash{}

	copy(r[:], v)

	return r, nil
}

// MustDecodeHashHex decodes the input str to [plumbing.Hash] and
// panics if any error is encountered.
func MustDecodeHashHex(str string) plumbing.Hash {
	v, err := DecodeHashHex(str)
	if err != nil {
		panic(err)
	}

	return v
}

// DecodeHashHexes calls [DecodeHashHex] on a list of input strings.
func DecodeHashHexes(strs ...string) ([]plumbing.Hash, error) {
	result := make([]plumbing.Hash, 0, len(strs))

	for _, v := range strs {
		x, err := DecodeHashHex(v)
		if err != nil {
			return nil, err
		}

		result = append(result, x)
	}

	return result, nil
}

// MustDecodeHashHexes decodes a list of input strings and panics if any error
// is encountered.
func MustDecodeHashHexes(strs ...string) []plumbing.Hash {
	r, err := DecodeHashHexes(strs...)
	if err != nil {
		panic(err)
	}

	return r
}
