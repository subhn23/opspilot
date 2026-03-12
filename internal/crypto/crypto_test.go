package crypto

import (
	"encoding/base64"
	"os"
	"testing"
)

func TestEncryptDecrypt(t *testing.T) {
	os.Setenv("ENCRYPTION_KEY", "0123456789abcdef0123456789abcdef") // 32 chars for AES-256
	defer os.Unsetenv("ENCRYPTION_KEY")

	originalText := "my-secret-ssh-key"
	
	encrypted, err := Encrypt(originalText)
	if err != nil {
		t.Fatalf("Encryption failed: %v", err)
	}

	if encrypted == originalText {
		t.Error("Encrypted text matches original text")
	}

	decrypted, err := Decrypt(encrypted)
	if err != nil {
		t.Fatalf("Decryption failed: %v", err)
	}

	if decrypted != originalText {
		t.Errorf("Expected %s, got %s", originalText, decrypted)
	}
}

func TestCryptoErrorCases(t *testing.T) {
	// 1. Missing ENCRYPTION_KEY
	os.Unsetenv("ENCRYPTION_KEY")
	_, err := Encrypt("test")
	if err == nil || err.Error() != "ENCRYPTION_KEY not set" {
		t.Error("Expected error for missing ENCRYPTION_KEY in Encrypt")
	}

	_, err = Decrypt("test")
	if err == nil || err.Error() != "ENCRYPTION_KEY not set" {
		t.Error("Expected error for missing ENCRYPTION_KEY in Decrypt")
	}

	// 2. Invalid Key Length
	os.Setenv("ENCRYPTION_KEY", "short")
	_, err = Encrypt("test")
	if err == nil {
		t.Error("Expected error for short ENCRYPTION_KEY")
	}

	// 3. Invalid Ciphertext (not base64)
	os.Setenv("ENCRYPTION_KEY", "0123456789abcdef0123456789abcdef")
	_, err = Decrypt("!!!")
	if err == nil {
		t.Error("Expected error for non-base64 input in Decrypt")
	}

	// 4. Short Ciphertext
	_, err = Decrypt(base64.StdEncoding.EncodeToString([]byte("short")))
	if err == nil || err.Error() != "ciphertext too short" {
		t.Error("Expected error for too short ciphertext")
	}
}
