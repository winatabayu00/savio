package password

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

// Argon2id parameters. Moderate, CI-friendly settings based on RFC 9106 guidance.
const (
	argonTime    = 1
	argonMemory  = 64 * 1024
	argonThreads = 2
	argonKeyLen  = 32
	saltLen      = 16
)

var ErrInvalidHashing = errors.New("invalid password hash")

// Hash returns a PHC-format Argon2id hash: $argon2id$v=19$m=...,t=...,p=...$salt$key
func Hash(password string) (string, error) {
	salt := make([]byte, saltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	key := argon2.IDKey([]byte(password), salt, argonTime, argonMemory, argonThreads, argonKeyLen)
	return encodePHC(salt, key), nil
}

// Verify checks a password against an Argon2id PHC hash. Uses constant-time
// comparison on the derived key.
func Verify(password, encoded string) (bool, error) {
	params, salt, key, err := decodePHC(encoded)
	if err != nil {
		return false, err
	}
	other := argon2.IDKey([]byte(password), salt, params.time, params.memory, uint8(params.threads), params.keyLen)
	return subtle.ConstantTimeCompare(key, other) == 1, nil
}

type phcParams struct {
	time, memory, threads, keyLen uint32
}

func encodePHC(salt, key []byte) string {
	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		19, argonMemory, argonTime, argonThreads,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(key))
}

func decodePHC(encoded string) (phcParams, []byte, []byte, error) {
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 || parts[1] != "argon2id" {
		return phcParams{}, nil, nil, ErrInvalidHashing
	}
	var v int
	if _, err := fmt.Sscanf(parts[2], "v=%d", &v); err != nil || v != 19 {
		return phcParams{}, nil, nil, ErrInvalidHashing
	}
	var m, t, p uint32
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &m, &t, &p); err != nil {
		return phcParams{}, nil, nil, ErrInvalidHashing
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return phcParams{}, nil, nil, ErrInvalidHashing
	}
	key, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return phcParams{}, nil, nil, ErrInvalidHashing
	}
	return phcParams{memory: m, time: t, threads: p, keyLen: uint32(len(key))}, salt, key, nil
}
