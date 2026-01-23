package storage

import (
	"fmt"
	"time"
)

// ConversationState captures minimal information about an active conversation
// so we can resume metadata across reconnects.
type ConversationState struct {
	Key       string `gorm:"primaryKey"`
	ChatID    int64  `gorm:"index"`
	UserID    int64  `gorm:"index"`
	Step      string
	Payload   []byte
	ExpiresAt time.Time `gorm:"index"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

// ConversationKey builds the storage key for a chat/user pair.
func ConversationKey(chatID, userID int64) string {
	return fmt.Sprintf("%d:%d", chatID, userID)
}

// SaveConversationState persists the given state. No-op for in-memory storage.
func (p *PeerStorage) SaveConversationState(state *ConversationState) error {
	if p == nil || p.inMemory || p.SqlSession == nil {
		return nil
	}
	if state.Key == "" {
		state.Key = ConversationKey(state.ChatID, state.UserID)
	}
	return p.SqlSession.Save(state).Error
}

// DeleteConversationState removes a state entry by key. No-op for in-memory storage.
func (p *PeerStorage) DeleteConversationState(key string) error {
	if p == nil || p.inMemory || p.SqlSession == nil {
		return nil
	}
	return p.SqlSession.Delete(&ConversationState{Key: key}).Error
}

// LoadConversationState retrieves a state entry by key.
func (p *PeerStorage) LoadConversationState(key string) (*ConversationState, error) {
	if p == nil || p.inMemory || p.SqlSession == nil {
		return nil, nil
	}
	var state ConversationState
	if err := p.SqlSession.Where("key = ?", key).First(&state).Error; err != nil {
		return nil, err
	}
	return &state, nil
}

// ListConversationStates fetches all persisted conversation states.
func (p *PeerStorage) ListConversationStates() ([]ConversationState, error) {
	if p == nil || p.inMemory || p.SqlSession == nil {
		return nil, nil
	}
	var states []ConversationState
	if err := p.SqlSession.Find(&states).Error; err != nil {
		return nil, err
	}
	return states, nil
}
