package main

import (
	"crypto/rand"
	"time"
)

// ULID implementation - Universally Unique Lexicographically Sortable Identifier
// Format: 26 characters, base32 encoded
// First 10 chars: timestamp (48 bits)
// Last 16 chars: randomness (80 bits)

const (
	ulidEncoding = "0123456789ABCDEFGHJKMNPQRSTVWXYZ" // Crockford's Base32
	ulidLength   = 26
)

var (
	encodeMap = []byte(ulidEncoding)
	decodeMap [256]byte
)

func init() {
	for i := range decodeMap {
		decodeMap[i] = 0xFF
	}
	for i, c := range encodeMap {
		decodeMap[c] = byte(i)
	}
}

// GenerateULID generates a new ULID with the given prefix
func GenerateULID(prefix string) string {
	// Get current timestamp in milliseconds
	now := time.Now()
	ms := uint64(now.UnixNano() / 1e6)

	// Encode timestamp (48 bits = 10 base32 chars)
	var result [ulidLength]byte
	for i := 9; i >= 0; i-- {
		result[i] = encodeMap[ms&0x1F]
		ms >>= 5
	}

	// Generate random bytes (80 bits = 10 bytes = 16 base32 chars)
	randomBytes := make([]byte, 10)
	if _, err := rand.Read(randomBytes); err != nil {
		panic(err)
	}

	// Encode random part
	var random uint64
	for i := 0; i < 10; i++ {
		random = (random << 8) | uint64(randomBytes[i])
		if i == 4 {
			// First 5 bytes encoded
			for j := 15; j >= 10; j-- {
				result[j] = encodeMap[random&0x1F]
				random >>= 5
			}
			random = 0
		}
	}
	// Last 5 bytes encoded
	for j := 25; j >= 16; j-- {
		result[j] = encodeMap[random&0x1F]
		random >>= 5
	}

	return prefix + "_" + string(result[:])
}

// GenerateEventID generates an event ID (just a ULID without prefix)
func GenerateEventID() string {
	return GenerateULID("")[1:] // Remove the leading underscore
}

// Simple lamport timestamp counter
var lamportCounter int64 = 0

// GetNextLamportTS returns the next Lamport timestamp
func GetNextLamportTS() int64 {
	lamportCounter++
	return lamportCounter
}
