package functions

import (
	"context"
	"fmt"
	"strings"

	"github.com/gotd/td/tg"
	mtp_errors "github.com/pageton/gotg/errors"
	"github.com/pageton/gotg/storage"
)

// GetUserByID fetches a user by ID from Telegram API.
//
// Parameters:
//   - ctx: Context for the API call
//   - raw: The raw Telegram client
//   - p: Peer storage for resolving peer references
//   - userID: The user ID to fetch
//
// Returns user object or an error.
func GetUserByID(ctx context.Context, raw *tg.Client, p *storage.PeerStorage, userID int64) (*tg.User, error) {
	inputPeer := GetInputPeerClassFromID(p, userID)
	if inputPeer == nil {
		return nil, mtp_errors.ErrPeerNotFound
	}

	switch peer := inputPeer.(type) {
	case *tg.InputPeerUser:
		users, err := raw.UsersGetUsers(ctx, []tg.InputUserClass{
			&tg.InputUser{
				UserID:     peer.UserID,
				AccessHash: peer.AccessHash,
			},
		})
		if err != nil {
			return nil, err
		}
		if len(users) == 0 {
			return nil, mtp_errors.ErrPeerNotFound
		}
		user, ok := users[0].(*tg.User)
		if !ok {
			return nil, mtp_errors.ErrPeerNotFound
		}
		return user, nil
	default:
		return nil, mtp_errors.ErrNotUser
	}
}

const (
	Admin      = "admin"
	Creator    = "creator"
	Member     = "member"
	Restricted = "restricted"
	Left       = "left"
)

// GetChatMember fetches information about a chat member.
// For channels, returns tg.ChannelParticipantClass with member details.
// For regular chats, use GetChatMemberInChat instead.
//
// Parameters:
//   - ctx: Context for the API call
//   - raw: The raw Telegram client
//   - p: Peer storage for resolving peer references
//   - chatID: The channel ID to query
//   - userID: The user ID to look up
//
// Returns participant information or an error.
func GetChatMember(ctx context.Context, raw *tg.Client, p *storage.PeerStorage, chatID, userID int64) (tg.ChannelParticipantClass, error) {
	inputPeer := GetInputPeerClassFromID(p, chatID)
	if inputPeer == nil {
		return nil, mtp_errors.ErrPeerNotFound
	}

	switch peer := inputPeer.(type) {
	case *tg.InputPeerChannel:
		res, err := raw.ChannelsGetParticipant(ctx, &tg.ChannelsGetParticipantRequest{
			Channel: &tg.InputChannel{
				ChannelID:  peer.ChannelID,
				AccessHash: peer.AccessHash,
			},
			Participant: GetInputPeerClassFromID(p, userID),
		})
		if err != nil {
			// Handle USER_NOT_PARTICIPANT error - return left status
			if strings.Contains(err.Error(), tg.ErrUserNotParticipant) {
				return &tg.ChannelParticipantLeft{
					Peer: &tg.PeerUser{UserID: userID},
				}, nil
			}
			return nil, err
		}
		SavePeersFromClassArray(p, res.Chats, res.Users)
		return res.Participant, nil
	case *tg.InputPeerChat, *tg.InputPeerUser:
		return nil, fmt.Errorf("get chat member for chats not implemented yet")
	default:
		return nil, mtp_errors.ErrNotChat
	}
}

// ExtractParticipantRights extracts admin rights from a channel participant.
// Returns nil if the participant is not an admin/creator or has no rights.
//
// Parameters:
//   - participant: The channel participant to extract rights from
//
// Returns ChatAdminRights if present, nil otherwise.
func ExtractParticipantRights(participant tg.ChannelParticipantClass) *tg.ChatAdminRights {
	switch p := participant.(type) {
	case *tg.ChannelParticipantCreator:
		return &p.AdminRights
	case *tg.ChannelParticipantAdmin:
		return &p.AdminRights
	case *tg.ChannelParticipant, *tg.ChannelParticipantBanned, *tg.ChannelParticipantLeft:
		return nil
	default:
		return nil
	}
}

// ExtractParticipantStatus extracts the status string from a channel participant.
//
// Parameters:
//   - participant: The channel participant to extract status from
//
// Returns status string (Creator, Admin, Member, Restricted, or Left).
func ExtractParticipantStatus(participant tg.ChannelParticipantClass) string {
	switch participant.(type) {
	case *tg.ChannelParticipantCreator:
		return Creator
	case *tg.ChannelParticipantAdmin:
		return Admin
	case *tg.ChannelParticipant, *tg.ChannelParticipantSelf:
		return Member
	case *tg.ChannelParticipantBanned:
		return Restricted
	case *tg.ChannelParticipantLeft:
		return Left
	default:
		return ""
	}
}

// ExtractParticipantRank extracts the rank string from a channel participant.
// Returns empty string for participants without rank.
//
// Parameters:
//   - participant: The channel participant to extract rank from
//
// Returns rank string if present, empty otherwise.
func ExtractParticipantTitle(participant tg.ChannelParticipantClass) string {
	switch p := participant.(type) {
	case *tg.ChannelParticipantCreator:
		return p.Rank
	case *tg.ChannelParticipantAdmin:
		return p.Rank
	default:
		return ""
	}
}

// ExtractParticipantUserID extracts the user ID from a channel participant.
//
// Parameters:
//   - participant: The channel participant to extract user ID from
//
// Returns user ID, or 0 if not found.
func ExtractParticipantUserID(participant tg.ChannelParticipantClass) int64 {
	switch p := participant.(type) {
	case *tg.ChannelParticipantCreator:
		return p.UserID
	case *tg.ChannelParticipantAdmin:
		return p.UserID
	case *tg.ChannelParticipant:
		return p.UserID
	case *tg.ChannelParticipantSelf:
		return p.UserID
	case *tg.ChannelParticipantBanned:
		if peer, ok := p.Peer.(*tg.PeerUser); ok {
			return peer.UserID
		}
	case *tg.ChannelParticipantLeft:
		if peer, ok := p.Peer.(*tg.PeerUser); ok {
			return peer.UserID
		}
	}
	return 0
}
