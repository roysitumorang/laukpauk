package helper

import (
	"bytes"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"runtime"

	"github.com/roysitumorang/laukpauk/pools"
	"golang.org/x/crypto/argon2"
)

var ErrInvalidHash = errors.New("argon2id: hash is not in the correct format")
var ErrIncompatibleVariant = errors.New("argon2id: incompatible variant of argon2")
var ErrIncompatibleVersion = errors.New("argon2id: incompatible version of argon2")
var DefaultArgon2Params = &Argon2Params{
	SaltLength: 16,
	KeyLength:  32,
	Iterations: 1,
	Memory:     64 * 1024,
	Threads:    uint8(runtime.NumCPU()),
}

type Argon2Params struct {
	// length of salt in bytes
	SaltLength uint32

	// length of derived key
	KeyLength uint32

	// number of iterations
	Iterations uint32

	// memory cost in KiB
	Memory uint32

	// number of parallel threads
	Threads uint8
}

func HashPassword(password string, params ...*Argon2Params) (*string, error) {
	sb := pools.BuilderPool.Get()
	defer func() {
		sb.Reset()
		pools.BuilderPool.Put(sb)
	}()
	sb.Reset()
	param := DefaultArgon2Params
	if len(params) > 0 {
		param = params[0]
	}
	salt := make([]byte, param.SaltLength)
	_, err := rand.Read(salt)
	if err != nil {
		return nil, err
	}
	key := argon2.IDKey(
		String2ByteSlice(password),
		salt,
		param.Iterations,
		param.Memory,
		param.Threads,
		param.KeyLength,
	)
	b64Enc := base64.RawStdEncoding
	_, _ = fmt.Fprintf(
		sb,
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		param.Memory,
		param.Iterations,
		param.Threads,
		b64Enc.EncodeToString(salt),
		b64Enc.EncodeToString(key),
	)
	return new(sb.String()), nil
}

func VerifyPassword(password, hash string) (bool, error) {
	bb := pools.BufferPool.Get()
	defer func() {
		bb.Reset()
		pools.BufferPool.Put(bb)
	}()
	bb.Reset()
	_, _ = bb.WriteString(hash)
	_, err := fmt.Fscanf(bb, "$argon2id$")
	if err != nil {
		return false, ErrIncompatibleVariant
	}
	var version int
	if _, err = fmt.Fscanf(bb, "v=%d$", &version); err != nil {
		return false, err
	}
	if version != argon2.Version {
		return false, ErrIncompatibleVersion
	}
	params := &Argon2Params{}
	if _, err = fmt.Fscanf(
		bb,
		"m=%d,t=%d,p=%d$",
		&params.Memory,
		&params.Iterations,
		&params.Threads,
	); err != nil {
		return false, err
	}
	rest, err := io.ReadAll(bb)
	if err != nil {
		return false, err
	}
	if bytes.ContainsAny(rest, "\r\n") {
		return false, ErrInvalidHash
	}
	i := bytes.IndexByte(rest, '$')
	if i == -1 {
		return false, ErrInvalidHash
	}
	b64Enc := base64.RawStdEncoding.Strict()
	salt := make([]byte, b64Enc.DecodedLen(i))
	_, err = b64Enc.Decode(salt, rest[:i])
	if err != nil {
		return false, err
	}
	params.SaltLength = uint32(len(salt))
	key := make([]byte, b64Enc.DecodedLen(len(rest)-i-1))
	if _, err = b64Enc.Decode(key, rest[i+1:]); err != nil {
		return false, err
	}
	params.KeyLength = uint32(len(key))
	otherKey := argon2.IDKey(String2ByteSlice(password), salt, params.Iterations, params.Memory, params.Threads, params.KeyLength)
	keyLen := int32(len(key))
	otherKeyLen := int32(len(otherKey))
	return subtle.ConstantTimeEq(keyLen, otherKeyLen) == 1 &&
		subtle.ConstantTimeCompare(key, otherKey) == 1, nil
}
