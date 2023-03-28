package test

import (
	"assignment3/aes"
	"assignment3/rsa"
	"bytes"
	"encoding/hex"
	"math/big"
	"testing"
)

// NOT ALL TEST GO THROUGH IF THEY ARE RUN AT THE SAME TIME. WHEN RUNNING EACH TEST
// INDIVIDUALLY, THEY ALL PASS

func TestIntegration(t *testing.T) {
	aes_key, _ := hex.DecodeString("6368616e676520746869732070617373")

	rsa_cipher := rsa.RSA(2048)
	rsa_secret_key := rsa_cipher.Keys.D

	aes_cipher := aes.AES(aes_key)
	aes_cipher.EncryptToFile("encrypt/integration.txt", rsa_secret_key.Bytes())

	rsa_secret_key_bytes := aes_cipher.DecryptFromFile("encrypt/integration.txt")
	rsa_secret_key = new(big.Int).SetBytes(rsa_secret_key_bytes)
	rsa_cipher.Keys.D = rsa_secret_key

	plaintext_before := []byte("hej")
	encrypted_rsa := rsa_cipher.Encrypt(plaintext_before)
	plaintext_after := rsa_cipher.Decrypt(encrypted_rsa)

	if !bytes.Equal(plaintext_before, plaintext_after) {
		t.Errorf("Plaintext = %d Decrypted %d", plaintext_before, plaintext_after)
	}
}
