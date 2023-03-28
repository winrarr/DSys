package test

import (
	"assignment3/aes"
	"assignment3/rsa"
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	random "math/rand"
	"testing"
	"time"
)

// This test does not pass when we run all test, but it passes
// when we run it individually (and we believe it should pass)
func TestRSA(t *testing.T) {
	for i := 0; i < 50; i++ {
		keyLength := random.Intn(2000) + 48
		rsa := rsa.RSA(keyLength)

		if rsa.Keys.N.BitLen() != keyLength {
			t.Errorf("Bad length of n")
		}

		before := []byte("hej")
		encrypted := rsa.Encrypt(before)
		after := rsa.Decrypt(encrypted)

		if !bytes.Equal(before, after) {
			t.Errorf("Plaintext = %d Decrypted = %d", before, after)
		}
	}
}

func TestAES(t *testing.T) {
	key, _ := hex.DecodeString("6368616e676520746869732070617373")

	aes := aes.AES(key)

	number := "12312335792"
	aes.EncryptToFile("encrypt/numbers.txt", []byte(number))
	decrypted_number := string(aes.DecryptFromFile("encrypt/numbers.txt"))

	str := "my name is bob"
	aes.EncryptToFile("encrypt/strings.txt", []byte(str))
	decrypted_string := string(aes.DecryptFromFile("encrypt/strings.txt"))

	if decrypted_number != number || decrypted_string != str {
		t.Errorf("Failed decryption...")
	}
}

// Test 1
func TestSignature(t *testing.T) {
	for i := 0; i < 20; i++ {
		rsa := rsa.RSA(2048)

		message := make([]byte, random.Intn(5000)+1000)
		rand.Read(message)

		signed_message := rsa.Sign(message)
		encrypted_signed_message := rsa.Encrypt(signed_message)

		decrypted_signed_message := rsa.Decrypt(encrypted_signed_message)
		verified, decrypted_message := rsa.VerifySignature(decrypted_signed_message)

		if !verified {
			t.Errorf("Bad signature")
			break
		}

		if !bytes.Equal(decrypted_message, message) {
			t.Errorf("Bad decryption")
			break
		}

		randombytes := make([]byte, 20)
		rand.Read(randombytes)
		decrypted_signed_message = append(randombytes, decrypted_signed_message[20:]...)

		verified, decrypted_message = rsa.VerifySignature(decrypted_signed_message)

		if verified {
			t.Errorf("Vefified bad signature")
		}

		if bytes.Equal(decrypted_message, message) {
			t.Errorf("Bytes should not be equal")
		}
	}
}

// Test 2
func TestHashTime(t *testing.T) {
	var totalBits float64 = 0
	var totalSeconds float64 = 0

	i := 0
	for ; i < 10000; i++ {
		length := random.Intn(10000) + 5000
		message := make([]byte, length)
		rand.Read(message)

		start := time.Now()
		sha256.Sum256(message)
		duration := time.Since(start)

		totalBits += float64(length * 8)
		totalSeconds += duration.Seconds()
	}

	fmt.Println(totalBits / totalSeconds)
}

// Test 3
func TestRSAsignTime(t *testing.T) {
	var totalTime float64 = 0

	i := 0
	for ; i < 10; i++ {
		rsa := rsa.RSA(2000)

		length := random.Intn(10000) + 5000
		message := make([]byte, length)
		rand.Read(message)

		hash := sha256.Sum256(message)

		start := time.Now()
		rsa.GetSignature(hash[:])
		duration := time.Since(start).Seconds()

		totalTime += float64(duration)
	}

	fmt.Println(totalTime / float64(i))
}

// Test 4 (Actual sign time in our implementation)
// 137ms versus our calculated 140ms
func TestSignTime(t *testing.T) {
	var totalTime float64 = 0

	i := 0
	for ; i < 10; i++ {
		rsa := rsa.RSA(2000)

		length := 10000
		message := make([]byte, length)
		rand.Read(message)

		start := time.Now()
		for i := 250; i+250 <= len(message); i += 250 {
			rsa.DecryptBlock(message[i : i+250])
		}
		totalTime += time.Since(start).Seconds()
	}

	fmt.Println(totalTime / float64(i))
}
