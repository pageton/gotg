package functions

import (
	"context"
	"fmt"
	"strings"

	"github.com/gotd/td/tg"
	"github.com/pageton/gotg/storage"
)

// ExportInvoice exports an invoice for use with payment providers.
//
// Parameters:
//   - ctx: Context for the API call
//   - raw: The raw Telegram client
//   - inputMedia: The invoice media to export
//
// Returns exported invoice or an error.
func ExportInvoice(ctx context.Context, raw *tg.Client, inputMedia tg.InputMediaClass) (*tg.PaymentsExportedInvoice, error) {
	return raw.PaymentsExportInvoice(ctx, inputMedia)
}

// SetPreCheckoutResults sets pre-checkout query results for a bot.
//
// Parameters:
//   - ctx: Context for the API call
//   - raw: The raw Telegram client
//   - success: Whether to checkout succeeded
//   - queryID: The pre-checkout query ID
//   - errMsg: Optional error message
//
// Returns true if successful, or an error.
func SetPreCheckoutResults(ctx context.Context, raw *tg.Client, success bool, queryID int64, errMsg string) (bool, error) {
	return raw.MessagesSetBotPrecheckoutResults(ctx, &tg.MessagesSetBotPrecheckoutResultsRequest{
		Success: success,
		QueryID: queryID,
		Error:   errMsg,
	})
}

// GetStarGifts retrieves the list of available star gifts.
//
// Parameters:
//   - ctx: Context for the API call
//   - raw: The raw Telegram client
//   - hash: Hash for caching (pass 0 to always fetch fresh)
//
// Returns the star gifts response or an error.
func GetStarGifts(ctx context.Context, raw *tg.Client, hash int) (tg.PaymentsStarGiftsClass, error) {
	return raw.PaymentsGetStarGifts(ctx, hash)
}

// GetResaleStarGifts retrieves the list of resale star gifts.
//
// Parameters:
//   - ctx: Context for the API call
//   - raw: The raw Telegram client
//   - req: The request parameters (sort, filter, pagination)
//
// Returns the resale star gifts response or an error.
func GetResaleStarGifts(ctx context.Context, raw *tg.Client, p *storage.PeerStorage, req *tg.PaymentsGetResaleStarGiftsRequest) (*tg.PaymentsResaleStarGifts, error) {
	result, err := raw.PaymentsGetResaleStarGifts(ctx, req)
	if err != nil {
		return nil, err
	}
	SavePeersFromClassArray(p, result.Chats, result.Users)
	return result, nil
}

// BuyResaleStarGift purchases a resale star gift by slug for a recipient.
//
// Parameters:
//   - ctx: Context for the API call
//   - raw: The raw Telegram client
//   - p: Peer storage for resolving peer references
//   - slug: The unique slug of the resale gift
//   - toID: Recipient as tg.InputPeerClass, int64 (userID), or string (username)
//
// Returns the payment result or an error.
func BuyResaleStarGift(ctx context.Context, raw *tg.Client, p *storage.PeerStorage, slug string, toID any) (tg.PaymentsPaymentResultClass, error) {
	peer, err := resolvePeer(ctx, raw, p, toID)
	if err != nil {
		return nil, fmt.Errorf("resolve recipient: %w", err)
	}

	inv := &tg.InputInvoiceStarGiftResale{
		Slug: slug,
		ToID: peer,
	}

	form, err := raw.PaymentsGetPaymentForm(ctx, &tg.PaymentsGetPaymentFormRequest{Invoice: inv})
	if err != nil {
		return nil, fmt.Errorf("get payment form: %w", err)
	}
	if f, ok := form.(*tg.PaymentsPaymentForm); ok {
		SavePeersFromClassArray(p, nil, f.Users)
	}
	if form.GetFormID() == 0 {
		return nil, fmt.Errorf("get payment form: empty form ID")
	}

	result, err := raw.PaymentsSendStarsForm(ctx, &tg.PaymentsSendStarsFormRequest{
		FormID:  form.GetFormID(),
		Invoice: inv,
	})
	if err != nil {
		return nil, fmt.Errorf("send stars form: %w", err)
	}

	return result, nil
}

func resolvePeer(ctx context.Context, raw *tg.Client, p *storage.PeerStorage, toID any) (tg.InputPeerClass, error) {
	switch v := toID.(type) {
	case tg.InputPeerClass:
		return v, nil
	case int64:
		peer := GetInputPeerClassFromID(p, v)
		if peer != nil {
			return peer, nil
		}
		return ResolveInputPeerByID(ctx, raw, p, v)
	case int:
		return resolvePeer(ctx, raw, p, int64(v))
	case string:
		username := strings.TrimPrefix(v, "@")
		if p != nil {
			if peer := p.GetPeerByUsername(username); peer != nil && peer.ID != 0 {
				return GetInputPeerClassFromID(p, peer.ID), nil
			}
		}
		resolved, err := raw.ContactsResolveUsername(ctx, &tg.ContactsResolveUsernameRequest{Username: username})
		if err != nil {
			return nil, err
		}
		SavePeersFromClassArray(p, resolved.Chats, resolved.Users)
		switch peer := resolved.Peer.(type) {
		case *tg.PeerUser:
			return &tg.InputPeerUser{UserID: peer.UserID, AccessHash: GetAccessHashFromResolved(resolved.Users, peer.UserID)}, nil
		case *tg.PeerChat:
			return &tg.InputPeerChat{ChatID: peer.ChatID}, nil
		case *tg.PeerChannel:
			return &tg.InputPeerChannel{ChannelID: peer.ChannelID, AccessHash: GetAccessHashFromResolvedChats(resolved.Chats, peer.ChannelID)}, nil
		}
		return nil, fmt.Errorf("could not resolve username %q", v)
	default:
		return nil, fmt.Errorf("unsupported toID type: %T", toID)
	}
}

// GetAccessHashFromResolved extracts access hash for a user from resolved users.
func GetAccessHashFromResolved(users []tg.UserClass, userID int64) int64 {
	for _, u := range users {
		if user, ok := u.(*tg.User); ok && user.ID == userID {
			return user.AccessHash
		}
	}
	return 0
}

// GetAccessHashFromResolvedChats extracts access hash for a channel from resolved chats.
func GetAccessHashFromResolvedChats(chats []tg.ChatClass, channelID int64) int64 {
	for _, c := range chats {
		if ch, ok := c.(*tg.Channel); ok && ch.ID == channelID {
			return ch.AccessHash
		}
	}
	return 0
}
