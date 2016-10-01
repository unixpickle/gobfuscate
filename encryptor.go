package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

// An Encrypter encrypts textual tokens.
type Encrypter struct {
	Key string
}

// Encrypt encrypts the token.
// The case of the first letter of the token is preserved.
func (e *Encrypter) Encrypt(token string) string {
	hashArray := sha256.Sum256([]byte(e.Key))
	key := hashArray[:]
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}
	iv := make([]byte, block.BlockSize())
	for i := range iv {
		iv[i] = byte(i)
	}
	enc := cipher.NewCFBEncrypter(block, iv)

	sourceBytes := []byte(token)
	resBytes := make([]byte, len(sourceBytes))

	enc.XORKeyStream(resBytes, sourceBytes)

	hexStr := strings.ToLower(hex.EncodeToString(resBytes))
	for i, x := range hexStr {
		if x >= '0' && x <= '9' {
			x = 'g' + (x - '0')
			hexStr = hexStr[:i] + string(x) + hexStr[i+1:]
		}
	}
	if strings.ToUpper(token[:1]) == token[:1] {
		hexStr = strings.ToUpper(hexStr[:1]) + hexStr[1:]
	}
	return hexStr
}
