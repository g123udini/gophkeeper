package aes

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"testing"
)

func decrypt(key, ciphertext []byte) ([]byte, error) {
	shaKey := sha256.Sum256(key)
	key = shaKey[:]
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, err
	}
	nonce, ct := ciphertext[:nonceSize], ciphertext[nonceSize:]
	return gcm.Open(nil, nonce, ct, nil)
}

func TestEncrypt_Success(t *testing.T) {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}
	data := []byte("test message")
	ciphertext, err := Encrypt(key, data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if bytes.Equal(data, ciphertext) {
		t.Errorf("ciphertext should not be equal to plaintext")
	}
	if len(ciphertext) <= len(data) {
		t.Errorf("ciphertext length is not greater than plaintext")
	}

	plaintext, err := decrypt(key, ciphertext)
	if err != nil {
		t.Fatalf("decryption failed: %v", err)
	}
	if !bytes.Equal(data, plaintext) {
		t.Errorf("decrypted plaintext does not match original, got %s, want %s", plaintext, data)
	}
}

func TestEncrypt_NilData(t *testing.T) {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}
	var data []byte
	ciphertext, err := Encrypt(key, data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ciphertext) == 0 {
		t.Errorf("ciphertext should not be empty")
	}
}
