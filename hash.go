package main

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

const hashedSymbolSize = 10

// A Padding is added to the input of a hash function
// to make it 'impossible' to find the input value
type Padding []byte

// Hash hashes the padding + token.
// The case of the first letter of the token is preserved.
func (p Padding) Hash(token string) string {
	hashArray := sha256.Sum256(append(p, []byte(token)...))

	hexStr := strings.ToLower(hex.EncodeToString(hashArray[:hashedSymbolSize]))
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
