package rsa

// Encrypting and decrypting a byte array starting with 0 removes the 0
// We make blocks of size blocksize-1 and pad a random byte at the beginning
// The padded byte cannot be larger than 127, otherwise we get overflow problems

func (c *Cipher) splitPlainBlocks(b []byte) ([][]byte, []byte) {
	blocks := make([][]byte, 0)

	i := 0
	for ; i+c.blockSize-1 <= len(b); i += c.blockSize - 1 {
		blocks = append(blocks, padStartOne(b[i:i+c.blockSize-1]))
	}

	remainder := padStartOne(b[i:])

	return blocks, remainder
}

func (c *Cipher) splitEncryptedBlocks(b []byte) ([][]byte, []byte) {
	blocks := make([][]byte, 0)

	i := 0
	for ; i+c.blockSize <= len(b); i += c.blockSize {
		blocks = append(blocks, b[i:i+c.blockSize])
	}

	remainder := b[i:]

	return blocks, remainder
}
