package types

import (
	"context"
	"fmt"

	"github.com/gotd/td/tg"
	"github.com/pageton/gotg/storage"
)

// InlineQuery represents a Telegram inline query with bound context for method chaining.
// It wraps tg.UpdateBotInlineQuery and provides convenience methods for common operations
// like answering the query and accessing user information.
type InlineQuery struct {
	*tg.UpdateBotInlineQuery
	// Context fields for bound methods
	Ctx         context.Context
	RawClient   *tg.Client
	PeerStorage *storage.PeerStorage
	SelfID      int64
	// Entities contains mapped users from the update for user lookup.
	Entities *tg.Entities
}

// ConstructInlineQuery creates an InlineQuery from tg.UpdateBotInlineQuery without context binding.
// The returned InlineQuery will have nil context fields and cannot perform bound operations.
// For full functionality, use ConstructInlineQueryWithContext instead.
func ConstructInlineQuery(iq *tg.UpdateBotInlineQuery) *InlineQuery {
	return ConstructInlineQueryWithContext(iq, nil, nil, nil, 0, nil)
}

// ConstructInlineQueryWithContext creates an InlineQuery with full context binding.
// This is the preferred constructor as it enables bound methods like Answer(), GetUser(), etc.
//
// Parameters:
//   - iq: The inline query update from Telegram
//   - ctx: Context for cancellation and timeouts
//   - raw: The raw Telegram client for API calls
//   - peerStorage: Peer storage for resolving peer references
//   - selfID: The current user/bot ID for context
//   - entities: Entities from the update for user lookup
func ConstructInlineQueryWithContext(iq *tg.UpdateBotInlineQuery, ctx context.Context, raw *tg.Client, peerStorage *storage.PeerStorage, selfID int64, entities *tg.Entities) *InlineQuery {
	if iq == nil {
		return nil
	}
	return &InlineQuery{
		UpdateBotInlineQuery: iq,
		Ctx:                  ctx,
		RawClient:            raw,
		PeerStorage:          peerStorage,
		SelfID:               selfID,
		Entities:             entities,
	}
}

// AnswerOpts contains optional parameters for answering inline queries.
type AnswerOpts struct {
	// CacheTime is the maximum time in seconds that the results may be cached on Telegram's servers.
	// Defaults to 300 (5 minutes) if not specified.
	CacheTime int
	// IsPersonal indicates whether the results should be cached only for the requesting user.
	// If true, results won't be shared with other users who send the same query.
	IsPersonal bool
	// NextOffset is the offset that will be passed to the next inline query when the user scrolls.
	// Pass an empty string if there are no more results or pagination is not supported.
	NextOffset string
	// SwitchPm instructs the client to display a button that switches the user to a private chat
	// with the bot and sends a start message.
	SwitchPm tg.InlineBotSwitchPM
	// SwitchWebview instructs the client to display a button that opens a Web App.
	SwitchWebview tg.InlineBotWebView
}

// Answer responds to the inline query with the provided results.
// This is the primary method for responding to inline queries.
//
// Parameters:
//   - results: A slice of inline query results to display to the user
//   - opts: Optional parameters for caching, pagination, and buttons (can be nil)
//
// Returns true if successful, or an error if the operation fails.
//
// Example:
//
//	results := []tg.InputBotInlineResultClass{
//	    &tg.InputBotInlineResult{
//	        ID:    "1",
//	        Type:  "article",
//	        Title: "Result 1",
//	        SendMessage: &tg.InputBotInlineMessageText{
//	            Message: "You selected result 1",
//	        },
//	    },
//	}
//	success, err := iq.Answer(results, &types.AnswerOpts{CacheTime: 60})
func (iq *InlineQuery) Answer(results []tg.InputBotInlineResultClass, opts *AnswerOpts) (bool, error) {
	if iq.RawClient == nil {
		return false, fmt.Errorf("inline query has no client context")
	}

	req := &tg.MessagesSetInlineBotResultsRequest{
		QueryID: iq.QueryID,
		Results: results,
	}

	if opts != nil {
		req.CacheTime = opts.CacheTime
		req.Private = opts.IsPersonal
		req.NextOffset = opts.NextOffset
		req.SwitchPm = opts.SwitchPm
		req.SwitchWebview = opts.SwitchWebview
	}

	return iq.RawClient.MessagesSetInlineBotResults(iq.Ctx, req)
}

// AnswerWithGallery responds to the inline query with results displayed in a gallery format.
// This is equivalent to Answer() but sets the gallery flag for photo/video results.
func (iq *InlineQuery) AnswerWithGallery(results []tg.InputBotInlineResultClass, opts *AnswerOpts) (bool, error) {
	if iq.RawClient == nil {
		return false, fmt.Errorf("inline query has no client context")
	}

	req := &tg.MessagesSetInlineBotResultsRequest{
		QueryID: iq.QueryID,
		Results: results,
		Gallery: true,
	}

	if opts != nil {
		req.CacheTime = opts.CacheTime
		req.Private = opts.IsPersonal
		req.NextOffset = opts.NextOffset
		req.SwitchPm = opts.SwitchPm
		req.SwitchWebview = opts.SwitchWebview
	}

	return iq.RawClient.MessagesSetInlineBotResults(iq.Ctx, req)
}

// GetUser returns the User who sent this inline query.
// Returns nil if the user cannot be found in entities or peer storage.
func (iq *InlineQuery) GetUser() *User {
	if iq.UserID == 0 {
		return nil
	}

	var tgUser *tg.User

	// Try to get user from entities first
	if iq.Entities != nil && iq.Entities.Users != nil {
		tgUser = iq.Entities.Users[iq.UserID]
	}

	// Fallback: try to construct minimal user from PeerStorage
	if tgUser == nil && iq.PeerStorage != nil {
		storagePeer := iq.PeerStorage.GetPeerByID(iq.UserID)
		if storagePeer != nil && storagePeer.Type == 1 { // 1 = TypeUser
			tgUser = &tg.User{
				ID:         iq.UserID,
				AccessHash: storagePeer.AccessHash,
			}
		}
	}

	// Final fallback: create stub user with just ID
	if tgUser == nil {
		tgUser = &tg.User{
			ID: iq.UserID,
		}
	}

	return &User{
		User:        *tgUser,
		Ctx:         iq.Ctx,
		RawClient:   iq.RawClient,
		PeerStorage: iq.PeerStorage,
		SelfID:      iq.SelfID,
	}
}

// Query returns the text of the inline query.
func (iq *InlineQuery) Query() string {
	return iq.UpdateBotInlineQuery.Query
}

// GetOffset returns the offset for pagination.
// This is the offset passed from a previous inline query result's NextOffset.
func (iq *InlineQuery) GetOffset() string {
	return iq.UpdateBotInlineQuery.Offset
}

// GetQueryID returns the unique identifier for this inline query.
// This ID is required when answering the query.
func (iq *InlineQuery) GetQueryID() int64 {
	return iq.QueryID
}

// GetUserID returns the ID of the user who sent this inline query.
func (iq *InlineQuery) GetUserID() int64 {
	return iq.UserID
}

// GetGeo returns the location of the user if they shared it, nil otherwise.
// Users can optionally share their location when making inline queries.
func (iq *InlineQuery) GetGeo() tg.GeoPointClass {
	return iq.Geo
}

// HasGeo returns true if the user shared their location with this query.
func (iq *InlineQuery) HasGeo() bool {
	return iq.Geo != nil
}

// GetPeerType returns the type of chat where the inline query originated.
// This can be used to customize results based on the chat context.
func (iq *InlineQuery) GetPeerType() tg.InlineQueryPeerTypeClass {
	return iq.PeerType
}

// IsFromPrivateChat returns true if the inline query was sent from a private chat.
func (iq *InlineQuery) IsFromPrivateChat() bool {
	_, ok := iq.PeerType.(*tg.InlineQueryPeerTypePM)
	return ok
}

// IsFromGroupChat returns true if the inline query was sent from a group chat.
func (iq *InlineQuery) IsFromGroupChat() bool {
	_, ok := iq.PeerType.(*tg.InlineQueryPeerTypeChat)
	return ok
}

// IsFromSupergroup returns true if the inline query was sent from a supergroup.
func (iq *InlineQuery) IsFromSupergroup() bool {
	_, ok := iq.PeerType.(*tg.InlineQueryPeerTypeMegagroup)
	return ok
}

// IsFromChannel returns true if the inline query was sent from a channel.
func (iq *InlineQuery) IsFromChannel() bool {
	_, ok := iq.PeerType.(*tg.InlineQueryPeerTypeBroadcast)
	return ok
}

// IsFromBotPM returns true if the inline query was sent from a private chat with the bot itself.
func (iq *InlineQuery) IsFromBotPM() bool {
	_, ok := iq.PeerType.(*tg.InlineQueryPeerTypeBotPM)
	return ok
}

// IsFromSameBotPM returns true if the inline query was sent from a private chat
// where the user is messaging the same bot.
func (iq *InlineQuery) IsFromSameBotPM() bool {
	_, ok := iq.PeerType.(*tg.InlineQueryPeerTypeSameBotPM)
	return ok
}

// Raw returns the underlying tg.UpdateBotInlineQuery struct.
func (iq *InlineQuery) Raw() *tg.UpdateBotInlineQuery {
	return iq.UpdateBotInlineQuery
}
