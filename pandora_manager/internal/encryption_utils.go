package internal

import (
	"crypto/rc4"
	"encoding/base64"
	"github.com/pkg/errors"
	"os"
)

func Encrypt(rawInput string) (string, error) {
	c, err := rc4.NewCipher(getKey())

	if err != nil {
		return "", errors.WithStack(err)
	}

	source := []byte(rawInput)
	dest := make([]byte, len(source))
	c.XORKeyStream(dest, source)

	return base64.RawStdEncoding.EncodeToString(dest), nil
}

func Decrypt(encryptedInput string) (string, error) {

	decode, err := base64.RawStdEncoding.DecodeString(encryptedInput)

	if err != nil {
		return "", errors.WithStack(err)
	}

	c, err := rc4.NewCipher(getKey())

	if err != nil {
		return "", errors.WithStack(err)
	}

	raw := make([]byte, len(decode))
	c.XORKeyStream(raw, decode)

	return string(raw), nil
}

func getKey() []byte {
	key := os.Getenv("encKey")

	if len(key) == 0 {
		panic("key should not be empty")
	}

	return []byte(key)
}
