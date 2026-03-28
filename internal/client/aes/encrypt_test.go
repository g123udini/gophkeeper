package aes

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"testing"

	"golang.org/x/crypto/argon2"
)

func decrypt(password, encrypted []byte) ([]byte, error) {
	if len(encrypted) < testSaltSize {
		return nil, errors.New("ciphertext too short")
	}

	salt := encrypted[:testSaltSize]
	key := argon2.IDKey(password, salt, 1, 64*1024, 4, 32)

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(encrypted) < testSaltSize+nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce := encrypted[testSaltSize : testSaltSize+nonceSize]
	ciphertext := encrypted[testSaltSize+nonceSize:]

	return gcm.Open(nil, nonce, ciphertext, nil)
}

func TestEncrypt_Success(t *testing.T) {
	password := make([]byte, 32)
	_, err := rand.Read(password)
	if err != nil {
		t.Fatalf("failed to generate password: %v", err)
	}

	data := []byte("test message")

	encrypted, err := Encrypt(password, data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if bytes.Equal(data, encrypted) {
		t.Errorf("encrypted data should not be equal to plaintext")
	}

	if len(encrypted) <= len(data) {
		t.Errorf("encrypted data length should be greater than plaintext length")
	}

	plaintext, err := decrypt(password, encrypted)
	if err != nil {
		t.Fatalf("decryption failed: %v", err)
	}

	if !bytes.Equal(data, plaintext) {
		t.Errorf("decrypted plaintext does not match original, got %s, want %s", plaintext, data)
	}
}

func TestEncrypt_NilData(t *testing.T) {
	password := make([]byte, 32)
	_, err := rand.Read(password)
	if err != nil {
		t.Fatalf("failed to generate password: %v", err)
	}

	var data []byte

	encrypted, err := Encrypt(password, data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(encrypted) == 0 {
		t.Errorf("encrypted data should not be empty")
	}

	plaintext, err := decrypt(password, encrypted)
	if err != nil {
		t.Fatalf("decryption failed: %v", err)
	}

	if len(plaintext) != 0 {
		t.Errorf("decrypted plaintext should be empty, got %q", plaintext)
	}
}
