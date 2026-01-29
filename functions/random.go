package functions

import (
	"math/rand"
	"sync"
	"time"
)

// Package-level shared random source and mutex to prevent memory leaks.
// This is initialized once when the package is loaded, preventing the creation
// of a new random source for each Context instance (Issue #112).
var (
	RandSource = rand.NewSource(time.Now().UnixNano())
	RandGen    = rand.New(RandSource)
	RandMutex  sync.Mutex
)

// GenerateRandomID generates a random int64 using the shared random source.
// This is thread-safe and prevents memory leaks (Issue #112).
func GenerateRandomID() int64 {
	RandMutex.Lock()
	defer RandMutex.Unlock()
	return RandGen.Int63()
}
