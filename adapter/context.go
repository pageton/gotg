package adapter

import (
	"github.com/pageton/gotg/functions"
)

// generateRandomID generates a random int64 for use in Telegram API calls.
// Random IDs are required by Telegram for duplicate request prevention.
// Uses thread-safe shared random source to prevent memory leaks (Issue #112).
func (ctx *Context) generateRandomID() int64 {
	return functions.GenerateRandomID()
}
