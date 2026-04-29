package broadcast

import (
	"fmt"
	"sync"
)

// ChatStatus represents the outcome of sending to a single chat.
type ChatStatus int

const (
	StatusPending ChatStatus = iota
	StatusSent
	StatusSkipped
	StatusFailed
)

func (s ChatStatus) String() string {
	switch s {
	case StatusSent:
		return "sent"
	case StatusSkipped:
		return "skipped"
	case StatusFailed:
		return "failed"
	default:
		return "pending"
	}
}

// BroadcastError records a per-chat failure with the error that caused it.
type BroadcastError struct {
	ChatID int64
	Err    error
}

func (e BroadcastError) Error() string {
	return fmt.Sprintf("broadcast: chat %d: %v", e.ChatID, e.Err)
}

func (e BroadcastError) Unwrap() error {
	return e.Err
}

// ChatOutcome is the result of attempting to deliver to a single chat.
type ChatOutcome struct {
	ChatID   int64
	Status   ChatStatus
	Attempts int
	Err      error
}

// BroadcastResult holds aggregate outcomes of a broadcast run.
// All mutation methods are safe for concurrent use.
type BroadcastResult struct {
	mu       sync.Mutex
	total    int
	sent     int
	skipped  int
	failed   int
	errors   []BroadcastError
	outcomes []ChatOutcome
}

// NewBroadcastResult creates a result pre-seeded with the total chat count.
func NewBroadcastResult(total int) *BroadcastResult {
	return &BroadcastResult{
		total:    total,
		errors:   make([]BroadcastError, 0),
		outcomes: make([]ChatOutcome, 0, total),
	}
}

func (r *BroadcastResult) Total() int   { return r.total }
func (r *BroadcastResult) Sent() int    { return r.sent }
func (r *BroadcastResult) Skipped() int { return r.skipped }
func (r *BroadcastResult) Failed() int  { return r.failed }
func (r *BroadcastResult) Errors() []BroadcastError {
	r.mu.Lock()
	defer r.mu.Unlock()
	cp := make([]BroadcastError, len(r.errors))
	copy(cp, r.errors)
	return cp
}

func (r *BroadcastResult) Outcomes() []ChatOutcome {
	r.mu.Lock()
	defer r.mu.Unlock()
	cp := make([]ChatOutcome, len(r.outcomes))
	copy(cp, r.outcomes)
	return cp
}

// RecordSent records a successful delivery to chatID.
func (r *BroadcastResult) RecordSent(chatID int64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.sent++
	r.outcomes = append(r.outcomes, ChatOutcome{ChatID: chatID, Status: StatusSent, Attempts: 1})
}

// RecordSkipped records a non-retryable skip for chatID (e.g. permission denied).
func (r *BroadcastResult) RecordSkipped(chatID int64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.skipped++
	r.outcomes = append(r.outcomes, ChatOutcome{ChatID: chatID, Status: StatusSkipped, Attempts: 1})
}

// RecordFailed records a failed delivery after exhausting retries.
func (r *BroadcastResult) RecordFailed(chatID int64, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.failed++
	r.errors = append(r.errors, BroadcastError{ChatID: chatID, Err: err})
	r.outcomes = append(r.outcomes, ChatOutcome{ChatID: chatID, Status: StatusFailed, Err: err})
}

// RecordFailedWithAttempts records a failed delivery with the number of attempts made.
func (r *BroadcastResult) RecordFailedWithAttempts(chatID int64, attempts int, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.failed++
	r.errors = append(r.errors, BroadcastError{ChatID: chatID, Err: err})
	r.outcomes = append(r.outcomes, ChatOutcome{ChatID: chatID, Status: StatusFailed, Attempts: attempts, Err: err})
}
