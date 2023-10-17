package gitrim

import "errors"

var ErrHexStringTooShort = errors.New("hex encoded byte slice is too short for hash")
