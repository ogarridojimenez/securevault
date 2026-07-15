package service

import "crypto/sha256"

func sha256Hash(data []byte) []byte {
	h := sha256.Sum256(data)
	return h[:]
}
