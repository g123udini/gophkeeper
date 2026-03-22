package aes

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"io"
	"testing"
)

func encrypt(key, plaintext []byte) ([]byte, error) {
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
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

func TestDecrypt_Success(t *testing.T) {
	key := make([]byte, 32) // AES-256
	if _, err := rand.Read(key); err != nil {
		t.Fatal(err)
	}
	plaintext := []byte("hello, world!")
	ciphertext, err := encrypt(key, plaintext)
	if err != nil {
		t.Fatalf("encryption failed: %v", err)
	}
	decrypted, err := Decrypt(key, ciphertext)
	if err != nil {
		t.Fatalf("decryption failed: %v", err)
	}
	if string(decrypted) != string(plaintext) {
		t.Errorf("decrypted != original. got %q, want %q", decrypted, plaintext)
	}
}

func TestDecrypt_WrongKey(t *testing.T) {
	key := make([]byte, 32)
	altKey := make([]byte, 32)
	rand.Read(key)
	rand.Read(altKey)
	plaintext := []byte("test data")
	ciphertext, err := encrypt(key, plaintext)
	if err != nil {
		t.Fatalf("encryption failed: %v", err)
	}
	_, err = Decrypt(altKey, ciphertext)
	if err == nil {
		t.Fatal("expected decryption to fail with wrong key, got no error")
	}
}

func TestDecrypt_CiphertextTooShort(t *testing.T) {
	key := make([]byte, 32)
	rand.Read(key)
	shortData := []byte("short")
	_, err := Decrypt(key, shortData)
	if err == nil {
		t.Fatal("expected ciphertext too short error")
	}
	if err.Error() != "ciphertext too short" {
		t.Error("expected ciphertext too short error")
	}
}

func TestDecrypt_InvalidCiphertext(t *testing.T) {
	key := make([]byte, 32)
	rand.Read(key)
	block, _ := aes.NewCipher(key)
	gcm, _ := cipher.NewGCM(block)
	nonce := make([]byte, gcm.NonceSize())
	rand.Read(nonce)
	invalidData := append(nonce, []byte("not a valid ciphertext")...)
	_, err := Decrypt(key, invalidData)
	if err == nil {
		t.Error("expected decryption to fail with invalid ciphertext")
	}
}
