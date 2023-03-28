package rsa

import (
	"bytes"
	"crypto/sha256"
)

func (c *Cipher) Sign(message []byte) []byte {
	signature := c.GetSignature(message)
	return append(message, padStartZeros(signature, c.blockSize)...)
}

func (c *Cipher) GetSignature(message []byte) []byte {
	plain_sha := sha256.Sum256(message)
	c.log(plain_sha[:], "[signing] plain sha")
	decrypted_sha := c.DecryptBlock(plain_sha[:])
	c.log(decrypted_sha, "[signing] signature (decrypted sha)")

	return decrypted_sha
}

// plain_sha has been through the notorious big.Int,
// so it has lost its starting zeros. Therefore we also
// remove the starting zeros from the calculated hash
func (c *Cipher) VerifySignature(m []byte) (bool, []byte) {
	message := m[:len(m)-c.blockSize]
	sha := sha256.Sum256(message)

	decrypted_sha := m[len(m)-c.blockSize:]
	c.log(decrypted_sha, "[verifying signature] signature (decrypted sha)")
	plain_sha := c.EncryptBlock(decrypted_sha)
	c.log(plain_sha, "[verifying signature] plain sha")

	return bytes.Equal(removeStartingZeros(sha[:]), plain_sha), message
}
