package storage

import (
	"fmt"
	"time"
)

type ConvState struct {
	Key       string
	ChatID    int64
	UserID    int64
	Step      string
	Payload   []byte
	ExpiresAt time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}

func ConvKey(chatID, userID int64) string {
	return fmt.Sprintf("%d:%d", chatID, userID)
}

func (p *PeerStorage) SaveConversationState(state *ConvState) error {
	if p == nil || p.inMemory || p.db == nil {
		return nil
	}
	if state.Key == "" {
		state.Key = ConvKey(state.ChatID, state.UserID)
	}
	return p.db.SaveConvState(state)
}

func (p *PeerStorage) DeleteConversationState(key string) error {
	if p == nil || p.inMemory || p.db == nil {
		return nil
	}
	return p.db.DeleteConvState(key)
}

func (p *PeerStorage) LoadConversationState(key string) (*ConvState, error) {
	if p == nil || p.inMemory || p.db == nil {
		return nil, nil
	}
	return p.db.LoadConvState(key)
}

func (p *PeerStorage) ListConversationStates() ([]ConvState, error) {
	if p == nil || p.inMemory || p.db == nil {
		return nil, nil
	}
	return p.db.ListConvStates()
}
