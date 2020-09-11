package internal

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	_ = os.Setenv("encKey", "test_key")

	code := m.Run()

	os.Exit(code)
}

func TestEncryptDecrypt(t *testing.T) {
	rawPassword := "abcds"
	encryptedPassword, err := Encrypt("abcds")

	if err != nil {
		t.Error(err)
		return
	}

	assert.Equal(t, "ieW1/Kk", encryptedPassword)

	decryptedPassword, err := Decrypt(encryptedPassword)

	if err != nil {
		t.Error(err)
		return
	}

	assert.Equal(t, rawPassword, decryptedPassword)
}
