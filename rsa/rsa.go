package rsa

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"math/big"
	random "math/rand"
)

type keys struct {
	E *big.Int
	N *big.Int
	D *big.Int
}

type Cipher struct {
	Keys      keys
	blockSize int
	Log       bool // Set this to true to get print statements
}

func RSA(k int) Cipher {
	e, n, d := keyGen(k)
	return Cipher{
		keys{
			E: e,
			N: n,
			D: d,
		},
		k / 8,
		false,
	}
}

// Encrypting and decrypting a byte array starting with 0 removes the 0
// We make blocks of size blocksize-1 and pad a random byte at the beginning
// The padded byte cannot be larger than 127, otherwise we get overflow problems

func (c *Cipher) Encrypt(m []byte) []byte {
	c.log(m, "[encrypting] starting")
	encrypted_bytes := make([]byte, 0)

	blocks, remainder := c.splitPlainBlocks(m)

	for _, block := range blocks {
		c.log(block, "[encrypting] plain block")
		encrypted_block := c.encryptAndPadBlock(block, c.blockSize)
		c.log(encrypted_block, "[encrypting] encrypted block")
		encrypted_bytes = append(encrypted_bytes, encrypted_block...)
		c.log(encrypted_bytes, "[encrypting] encrypted bytes")
	}

	if len(remainder) != 0 {
		c.log(remainder, "[encrypting] plain remainder block")
		encrypted_block := c.EncryptBlock(remainder)
		c.log(encrypted_block, "[encrypting] encrypted remainder block")
		encrypted_bytes = append(encrypted_bytes, encrypted_block...)
	}

	c.log(encrypted_bytes, "[encrypting] returning encrypted bytes")
	return encrypted_bytes
}

func (c *Cipher) Decrypt(m []byte) []byte {
	c.log(m, "[decrypting] starting")
	decrypted_bytes := make([]byte, 0)

	blocks, remainder := c.splitEncryptedBlocks(m)

	for _, block := range blocks {
		c.log(block, "[decrypting] encrypted block")
		decrypted_block := c.DecryptBlock(block)
		c.log(decrypted_block, "[decrypting] plain block")
		decrypted_bytes = append(decrypted_bytes, decrypted_block[1:]...)
		c.log(decrypted_bytes, "[decrypting] plain bytes")
	}

	if len(remainder) > 1 {
		c.log(remainder, "[decrypting] encrypted remainder block")
		decrypted_block := c.DecryptBlock(remainder)
		c.log(decrypted_block, "[decrypting] plain remainder block")
		decrypted_bytes = append(decrypted_bytes, decrypted_block[1:]...)
	}

	c.log(decrypted_bytes, "[decrypting] returning plain bytes")
	return decrypted_bytes
}

func (c *Cipher) encryptAndPadBlock(b []byte, length int) []byte {
	return padStartZeros(c.EncryptBlock(b), length)
}

func padStartZeros(b []byte, length int) []byte {
	return append(bytes.Repeat([]byte{0}, length-len(b)), b...)
}

func removeStartingZeros(b []byte) []byte {
	for i := range b {
		if b[i] != 0 {
			return b[i:]
		}
	}
	return b
}

func padStartOne(b []byte) []byte {
	return append([]byte{byte(random.Intn(127) + 1)}, b...)
}

func (c *Cipher) EncryptBlock(b []byte) []byte {
	i := new(big.Int).SetBytes(b)
	return i.Exp(i, c.Keys.E, c.Keys.N).Bytes()
}

func (c *Cipher) DecryptBlock(b []byte) []byte {
	i := new(big.Int).SetBytes(b)
	return i.Exp(i, c.Keys.D, c.Keys.N).Bytes()
}

func keyGen(k int) (*big.Int, *big.Int, *big.Int) {
	e := big.NewInt(3)
	p, p_sub_1 := createPrime(e, k/2)
	if k%2 == 1 { // if k is odd then we need a prime that is one bit longer
		k++
	}
	q, q_sub_1 := createPrime(e, k/2)

	pq := new(big.Int).Mul(p, q)
	pq_sub_1 := new(big.Int).Mul(p_sub_1, q_sub_1)

	d := new(big.Int).ModInverse(e, pq_sub_1)

	return e, pq, d
}

func createPrime(e *big.Int, k int) (*big.Int, *big.Int) {
	prime := new(big.Int)
	prime_sub_1 := new(big.Int)
	gcd := new(big.Int)

	for gcd.Cmp(big.NewInt(1)) != 0 {
		prime, _ = rand.Prime(rand.Reader, k)
		prime_sub_1.Sub(prime, big.NewInt(1))
		gcd.GCD(nil, nil, e, prime_sub_1)
	}

	return prime, prime_sub_1
}

func (c *Cipher) SetKeys(e *big.Int, n *big.Int, d *big.Int) {
	c.Keys.E = e
	c.Keys.N = n
	c.Keys.D = d
}

func (cipher *Cipher) PrintKeys() {
	fmt.Println(cipher.Keys.E.String())
	fmt.Println(cipher.Keys.N.String())
	fmt.Println(cipher.Keys.D.String())
}

func (c *Cipher) log(b []byte, msg string) {
	if !c.Log {
		return
	}
	println(msg)
	fmt.Println(b)
	println()
}
