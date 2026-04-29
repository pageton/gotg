package broadcast

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestChatOutcome_StatusString(t *testing.T) {
	tests := []struct {
		status   ChatStatus
		expected string
	}{
		{StatusSent, "sent"},
		{StatusSkipped, "skipped"},
		{StatusFailed, "failed"},
		{StatusPending, "pending"},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.expected, tt.status.String())
	}
}

func TestNewBroadcastResult(t *testing.T) {
	br := NewBroadcastResult(5)
	assert.Equal(t, 5, br.Total())
	assert.Equal(t, 0, br.Sent())
	assert.Equal(t, 0, br.Failed())
	assert.Equal(t, 0, br.Skipped())
	assert.Empty(t, br.Errors())
}

func TestBroadcastResult_RecordSent(t *testing.T) {
	br := NewBroadcastResult(3)
	br.RecordSent(100)
	br.RecordSent(200)
	assert.Equal(t, 2, br.Sent())
}

func TestBroadcastResult_RecordSkipped(t *testing.T) {
	br := NewBroadcastResult(3)
	br.RecordSkipped(100)
	assert.Equal(t, 1, br.Skipped())
}

func TestBroadcastResult_RecordFailed(t *testing.T) {
	br := NewBroadcastResult(3)
	br.RecordFailed(100, assert.AnError)
	br.RecordFailed(200, assert.AnError)
	assert.Equal(t, 2, br.Failed())
	assert.Len(t, br.Errors(), 2)
	assert.Equal(t, int64(100), br.Errors()[0].ChatID)
	assert.Equal(t, int64(200), br.Errors()[1].ChatID)
}

func TestBroadcastResult_ConcurrentRecords(t *testing.T) {
	br := NewBroadcastResult(100)
	var wg sync.WaitGroup
	for i := int64(0); i < 100; i++ {
		wg.Add(1)
		go func(id int64) {
			defer wg.Done()
			switch id % 3 {
			case 0:
				br.RecordSent(id)
			case 1:
				br.RecordSkipped(id)
			case 2:
				br.RecordFailed(id, assert.AnError)
			}
		}(i)
	}
	wg.Wait()

	// 34 sent (0,3,6,...,99), 33 skipped (1,4,7,...,97), 33 failed (2,5,8,...,98)
	assert.Equal(t, 34, br.Sent())
	assert.Equal(t, 33, br.Skipped())
	assert.Equal(t, 33, br.Failed())
	assert.Equal(t, 100, br.Sent()+br.Skipped()+br.Failed())
	assert.Len(t, br.Errors(), 33)
}

func TestBroadcastResult_Outcomes(t *testing.T) {
	br := NewBroadcastResult(2)
	br.RecordSent(100)
	br.RecordFailed(200, assert.AnError)
	outcomes := br.Outcomes()
	assert.Equal(t, 2, len(outcomes))
}
