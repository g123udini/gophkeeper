package aes

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io"
	"testing"

	"golang.org/x/crypto/argon2"
)

const testSaltSize = 16

func encrypt(password, plaintext []byte) ([]byte, error) {
	salt := make([]byte, testSaltSize)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, err
	}

	key := argon2.IDKey(password, salt, 1, 64*1024, 4, 32)

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

	ciphertext := gcm.Seal(nil, nonce, plaintext, nil)

	result := make([]byte, 0, len(salt)+len(nonce)+len(ciphertext))
	result = append(result, salt...)
	result = append(result, nonce...)
	result = append(result, ciphertext...)

	return result, nil
}

func TestDecrypt_Success(t *testing.T) {
	password := make([]byte, 32)
	if _, err := rand.Read(password); err != nil {
		t.Fatal(err)
	}

	plaintext := []byte("hello, world!")

	ciphertext, err := encrypt(password, plaintext)
	if err != nil {
		t.Fatalf("encryption failed: %v", err)
	}

	decrypted, err := Decrypt(password, ciphertext)
	if err != nil {
		t.Fatalf("decryption failed: %v", err)
	}

	if string(decrypted) != string(plaintext) {
		t.Errorf("decrypted != original. got %q, want %q", decrypted, plaintext)
	}
}

func TestDecrypt_WrongKey(t *testing.T) {
	password := make([]byte, 32)
	wrongPassword := make([]byte, 32)

	if _, err := rand.Read(password); err != nil {
		t.Fatal(err)
	}
	if _, err := rand.Read(wrongPassword); err != nil {
		t.Fatal(err)
	}

	plaintext := []byte("test data")

	ciphertext, err := encrypt(password, plaintext)
	if err != nil {
		t.Fatalf("encryption failed: %v", err)
	}

	_, err = Decrypt(wrongPassword, ciphertext)
	if err == nil {
		t.Fatal("expected decryption to fail with wrong key, got no error")
	}
}

func TestDecrypt_CiphertextTooShort(t *testing.T) {
	password := make([]byte, 32)
	if _, err := rand.Read(password); err != nil {
		t.Fatal(err)
	}

	shortData := []byte("short")

	_, err := Decrypt(password, shortData)
	if err == nil {
		t.Fatal("expected ciphertext too short error")
	}

	if !errors.Is(err, errors.New("ciphertext too short")) && err.Error() != "ciphertext too short" {
		t.Errorf("expected ciphertext too short error, got %v", err)
	}
}

func TestDecrypt_InvalidCiphertext(t *testing.T) {
	password := make([]byte, 32)
	if _, err := rand.Read(password); err != nil {
		t.Fatal(err)
	}

	salt := make([]byte, testSaltSize)
	if _, err := rand.Read(salt); err != nil {
		t.Fatal(err)
	}

	key := argon2.IDKey(password, salt, 1, 64*1024, 4, 32)

	block, err := aes.NewCipher(key)
	if err != nil {
		t.Fatal(err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		t.Fatal(err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		t.Fatal(err)
	}

	invalidData := make([]byte, 0, len(salt)+len(nonce)+len("not a valid ciphertext"))
	invalidData = append(invalidData, salt...)
	invalidData = append(invalidData, nonce...)
	invalidData = append(invalidData, []byte("not a valid ciphertext")...)

	_, err = Decrypt(password, invalidData)
	if err == nil {
		t.Error("expected decryption to fail with invalid ciphertext")
	}
}
