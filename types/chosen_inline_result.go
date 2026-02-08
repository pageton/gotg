package types

import (
	"context"

	"github.com/gotd/td/tg"
	"github.com/pageton/gotg/storage"
)

// ChosenInlineResult represents a chosen inline query result with bound context.
// It wraps tg.UpdateBotInlineSend and provides convenience methods for accessing
// user information and the inline message for editing.
type ChosenInlineResult struct {
	*tg.UpdateBotInlineSend
	Ctx         context.Context
	RawClient   *tg.Client
	PeerStorage *storage.PeerStorage
	SelfID      int64
	Entities    *tg.Entities
}

// ConstructChosenInlineResultWithContext creates a ChosenInlineResult with full context binding.
func ConstructChosenInlineResultWithContext(cir *tg.UpdateBotInlineSend, ctx context.Context, raw *tg.Client, peerStorage *storage.PeerStorage, selfID int64, entities *tg.Entities) *ChosenInlineResult {
	if cir == nil {
		return nil
	}
	return &ChosenInlineResult{
		UpdateBotInlineSend: cir,
		Ctx:                 ctx,
		RawClient:           raw,
		PeerStorage:         peerStorage,
		SelfID:              selfID,
		Entities:            entities,
	}
}

// GetUser returns the User who chose this inline result.
func (cir *ChosenInlineResult) GetUser() *User {
	if cir.UserID == 0 {
		return nil
	}

	var tgUser *tg.User

	if cir.Entities != nil && cir.Entities.Users != nil {
		tgUser = cir.Entities.Users[cir.UserID]
	}

	if tgUser == nil && cir.PeerStorage != nil {
		storagePeer := cir.PeerStorage.GetPeerByID(cir.UserID)
		if storagePeer != nil && storagePeer.Type == 1 {
			tgUser = &tg.User{
				ID:         cir.UserID,
				AccessHash: storagePeer.AccessHash,
			}
		}
	}

	if tgUser == nil {
		tgUser = &tg.User{ID: cir.UserID}
	}

	return &User{
		User:        *tgUser,
		Ctx:         cir.Ctx,
		RawClient:   cir.RawClient,
		PeerStorage: cir.PeerStorage,
		SelfID:      cir.SelfID,
	}
}

// ResultID returns the unique identifier of the chosen result.
func (cir *ChosenInlineResult) ResultID() string {
	return cir.ID
}

// Query returns the query that was used to obtain this result.
func (cir *ChosenInlineResult) Query() string {
	return cir.UpdateBotInlineSend.Query
}

// GetUserID returns the ID of the user who chose this result.
func (cir *ChosenInlineResult) GetUserID() int64 {
	return cir.UserID
}

// GetGeo returns the user's location if they shared it, nil otherwise.
func (cir *ChosenInlineResult) GetGeo() tg.GeoPointClass {
	return cir.Geo
}

// HasGeo returns true if the user shared their location.
func (cir *ChosenInlineResult) HasGeo() bool {
	return cir.Geo != nil
}

// GetInlineMessageID returns the inline message ID if available.
// This can be used to edit the sent inline message.
func (cir *ChosenInlineResult) GetInlineMessageID() tg.InputBotInlineMessageIDClass {
	msgID, ok := cir.UpdateBotInlineSend.GetMsgID()
	if !ok {
		return nil
	}
	return msgID
}

// HasInlineMessageID returns true if an inline message ID is available for editing.
func (cir *ChosenInlineResult) HasInlineMessageID() bool {
	_, ok := cir.UpdateBotInlineSend.GetMsgID()
	return ok
}

// Raw returns the underlying tg.UpdateBotInlineSend struct.
func (cir *ChosenInlineResult) Raw() *tg.UpdateBotInlineSend {
	return cir.UpdateBotInlineSend
}
