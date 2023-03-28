package aes

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"
	"io/ioutil"
	"log"
	"os"
)

/*
	A lot of code is heavily inspired by the example from the NewCTR function
*/

type Cipher struct {
	block cipher.Block
}

func AES(key []byte) Cipher {
	block, err := aes.NewCipher(key)
	if err != nil {
		log.Fatal(err)
	}

	return Cipher{
		block: block,
	}
}

func (c *Cipher) EncryptToFile(file string, plaintext []byte) {
	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		panic(err)
	}

	stream := cipher.NewCTR(c.block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)

	f, err := os.Create(file)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	err = ioutil.WriteFile(file, ciphertext, 0777)
	if err != nil {
		log.Fatal(err)
	}
}

func (c *Cipher) DecryptFromFile(encryptedFile string) []byte {
	file, err := os.Open(encryptedFile)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	ciphertext, err := ioutil.ReadAll(file)
	if err != nil {
		log.Fatal(err)
	}
	iv := ciphertext[:aes.BlockSize]
	plaintext := make([]byte, len(ciphertext)-aes.BlockSize)

	stream := cipher.NewCTR(c.block, iv)
	stream.XORKeyStream(plaintext, ciphertext[aes.BlockSize:])
	return plaintext
}
