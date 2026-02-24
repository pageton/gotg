package functions

import (
	"crypto/rand"
	"encoding/binary"
)

// GenerateRandomID generates a cryptographically random int64.
// This is thread-safe (crypto/rand is safe for concurrent use) and
// avoids the weak math/rand generator (Issue #112, gosec G404).
func GenerateRandomID() int64 {
	var buf [8]byte
	_, _ = rand.Read(buf[:]) // crypto/rand.Read never returns error on supported platforms
	return int64(binary.LittleEndian.Uint64(buf[:])) //nolint:gosec // G115: intentional uint64->int64 for random ID
}
