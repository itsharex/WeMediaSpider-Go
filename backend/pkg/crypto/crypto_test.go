package crypto

import (
	"strings"
	"testing"
)

func TestEncryptDecrypt_RoundTrip(t *testing.T) {
	plaintext := []byte("hello, WeMediaSpider!")
	password := "test-password-123"

	encrypted, err := Encrypt(plaintext, password)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}
	if encrypted == "" {
		t.Fatal("Encrypt returned empty string")
	}

	decrypted, err := Decrypt(encrypted, password)
	if err != nil {
		t.Fatalf("Decrypt failed: %v", err)
	}
	if string(decrypted) != string(plaintext) {
		t.Errorf("got %q, want %q", decrypted, plaintext)
	}
}

func TestEncrypt_ProducesUniqueCiphertexts(t *testing.T) {
	plaintext := []byte("same plaintext")
	password := "same-password"

	a, err := Encrypt(plaintext, password)
	if err != nil {
		t.Fatal(err)
	}
	b, err := Encrypt(plaintext, password)
	if err != nil {
		t.Fatal(err)
	}
	if a == b {
		t.Error("Encrypt should produce different ciphertexts due to random salt/nonce")
	}
}

func TestDecrypt_WrongPassword(t *testing.T) {
	encrypted, err := Encrypt([]byte("secret"), "correct-password")
	if err != nil {
		t.Fatal(err)
	}
	_, err = Decrypt(encrypted, "wrong-password")
	if err == nil {
		t.Error("Decrypt with wrong password should fail")
	}
}

func TestDecrypt_InvalidData(t *testing.T) {
	_, err := Decrypt("not-valid-base64!!!", "password")
	if err == nil {
		t.Error("Decrypt with invalid base64 should fail")
	}

	_, err = Decrypt("dG9vc2hvcnQ=", "password") // base64("tooshort")
	if err == nil {
		t.Error("Decrypt with too-short data should fail")
	}
}

func TestEncrypt_EmptyPlaintext(t *testing.T) {
	encrypted, err := Encrypt([]byte{}, "password")
	if err != nil {
		t.Fatalf("Encrypt of empty plaintext failed: %v", err)
	}
	decrypted, err := Decrypt(encrypted, "password")
	if err != nil {
		t.Fatalf("Decrypt of empty plaintext failed: %v", err)
	}
	if len(decrypted) != 0 {
		t.Errorf("expected empty plaintext, got %q", decrypted)
	}
}

func TestEncryptDecryptZGSWX_RoundTrip(t *testing.T) {
	payload := []byte(`{"token":"abc","cookies":[]}`)
	masterKey := make([]byte, 32)
	for i := range masterKey {
		masterKey[i] = byte(i + 1)
	}

	encrypted, err := EncryptToZGSWX(payload, masterKey)
	if err != nil {
		t.Fatalf("EncryptToZGSWX failed: %v", err)
	}

	if err := ValidateZGSWXFormat(encrypted); err != nil {
		t.Fatalf("ValidateZGSWXFormat failed: %v", err)
	}

	decrypted, err := DecryptFromZGSWX(encrypted, masterKey)
	if err != nil {
		t.Fatalf("DecryptFromZGSWX failed: %v", err)
	}
	if string(decrypted) != string(payload) {
		t.Errorf("got %q, want %q", decrypted, payload)
	}
}

func TestDecryptZGSWX_WrongKey(t *testing.T) {
	masterKey := make([]byte, 32)
	for i := range masterKey {
		masterKey[i] = byte(i + 1)
	}
	wrongKey := make([]byte, 32)

	encrypted, err := EncryptToZGSWX([]byte("secret"), masterKey)
	if err != nil {
		t.Fatal(err)
	}
	_, err = DecryptFromZGSWX(encrypted, wrongKey)
	if err == nil {
		t.Error("DecryptFromZGSWX with wrong key should fail")
	}
}

func TestValidateZGSWXFormat_InvalidMagic(t *testing.T) {
	bad := make([]byte, HeaderSize+1)
	err := ValidateZGSWXFormat(bad)
	if err == nil || !strings.Contains(err.Error(), "magic") {
		t.Errorf("expected magic number error, got %v", err)
	}
}
