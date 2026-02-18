package functions

import (
	"context"
	"fmt"

	"github.com/gotd/td/tg"
	"github.com/pageton/gotg/errors"
	"github.com/pageton/gotg/storage"
)

const (
	Admin      = "admin"
	Creator    = "creator"
	Member     = "member"
	Restricted = "restricted"
	Left       = "left"
)

// ChatMembersFilter specifies which types of members to return from GetChatMembers.
type ChatMembersFilter int

const (
	// FilterSearch returns members matching a search query (default).
	FilterSearch ChatMembersFilter = iota
	// FilterRecent returns recently active members.
	FilterRecent
	// FilterAdmins returns only administrators.
	FilterAdmins
	// FilterBots returns only bots.
	FilterBots
	// FilterKicked returns only kicked (banned) members.
	FilterKicked
	// FilterBanned returns only restricted members.
	FilterBanned
	// FilterContacts returns only contacts.
	FilterContacts
)

// toTL converts ChatMembersFilter to the corresponding tg.ChannelParticipantsFilterClass.
func (f ChatMembersFilter) toTL(query string) tg.ChannelParticipantsFilterClass {
	switch f {
	case FilterRecent:
		return &tg.ChannelParticipantsRecent{}
	case FilterAdmins:
		return &tg.ChannelParticipantsAdmins{}
	case FilterBots:
		return &tg.ChannelParticipantsBots{}
	case FilterKicked:
		return &tg.ChannelParticipantsKicked{Q: query}
	case FilterBanned:
		return &tg.ChannelParticipantsBanned{Q: query}
	case FilterContacts:
		return &tg.ChannelParticipantsContacts{Q: query}
	default: // FilterSearch
		return &tg.ChannelParticipantsSearch{Q: query}
	}
}

// GetChatMembersOpts holds optional parameters for GetChatMembers.
type GetChatMembersOpts struct {
	// Query filters members by display name or username.
	// Only applicable to FilterSearch, FilterKicked, and FilterBanned.
	Query string

	// Limit is the maximum number of members to return.
	// Defaults to 200. Maximum per-request is 200.
	Limit int

	// Filter selects which type of members to retrieve.
	// Defaults to FilterSearch.
	Filter ChatMembersFilter
}

// GetChatMembers returns a list of members from a chat or channel.
// For channels/supergroups, it calls channels.getParticipants with pagination.
// For basic groups, it fetches all participants via messages.getFullChat.
func GetChatMembers(ctx context.Context, raw *tg.Client, p *storage.PeerStorage, chatID int64, opts ...*GetChatMembersOpts) ([]tg.ChannelParticipantClass, error) {
	var opt GetChatMembersOpts
	if len(opts) > 0 && opts[0] != nil {
		opt = *opts[0]
	}
	if opt.Limit <= 0 {
		opt.Limit = 200
	}

	inputPeer := GetInputPeerClassFromID(p, chatID)
	if inputPeer == nil {
		// Fallback: try to resolve the peer via API (e.g. after participant updates
		// where the channel may not yet be in storage).
		var err error
		inputPeer, err = ResolveInputPeerByID(ctx, raw, p, chatID)
		if err != nil {
			return nil, err
		}
	}

	switch peer := inputPeer.(type) {
	case *tg.InputPeerChannel:
		return getChatMembersChannel(ctx, raw, p, peer, &opt)
	case *tg.InputPeerChat:
		return getChatMembersChat(ctx, raw, p, chatID)
	default:
		return nil, errors.ErrNotChat
	}
}

func getChatMembersChannel(ctx context.Context, raw *tg.Client, p *storage.PeerStorage, peer *tg.InputPeerChannel, opt *GetChatMembersOpts) ([]tg.ChannelParticipantClass, error) {
	channel := &tg.InputChannel{
		ChannelID:  peer.ChannelID,
		AccessHash: peer.AccessHash,
	}
	filter := opt.Filter.toTL(opt.Query)

	total := opt.Limit
	maxPerRequest := 200
	var all []tg.ChannelParticipantClass
	offset := 0

	for {
		remaining := total - len(all)
		if remaining <= 0 {
			break
		}
		limit := min(remaining, maxPerRequest)

		res, err := raw.ChannelsGetParticipants(ctx, &tg.ChannelsGetParticipantsRequest{
			Channel: channel,
			Filter:  filter,
			Offset:  offset,
			Limit:   limit,
			Hash:    0,
		})
		if err != nil {
			return nil, err
		}

		participants, ok := res.(*tg.ChannelsChannelParticipants)
		if !ok {
			break
		}

		SavePeersFromClassArray(p, participants.Chats, participants.Users)

		if len(participants.Participants) == 0 {
			break
		}

		all = append(all, participants.Participants...)
		offset += len(participants.Participants)
	}

	return all, nil
}

func getChatMembersChat(ctx context.Context, raw *tg.Client, p *storage.PeerStorage, chatID int64) ([]tg.ChannelParticipantClass, error) {
	fullChat, err := raw.MessagesGetFullChat(ctx, chatID)
	if err != nil {
		return nil, err
	}

	SavePeersFromClassArray(p, fullChat.Chats, fullChat.Users)

	chatFull, ok := fullChat.FullChat.(*tg.ChatFull)
	if !ok {
		return nil, fmt.Errorf("could not get full chat info")
	}

	cp, ok := chatFull.Participants.(*tg.ChatParticipants)
	if !ok {
		return nil, fmt.Errorf("participants info is forbidden")
	}

	result := make([]tg.ChannelParticipantClass, 0, len(cp.Participants))
	for _, member := range cp.Participants {
		result = append(result, chatParticipantToChannel(member))
	}

	return result, nil
}

func chatParticipantToChannel(p tg.ChatParticipantClass) tg.ChannelParticipantClass {
	switch m := p.(type) {
	case *tg.ChatParticipantCreator:
		return &tg.ChannelParticipantCreator{
			UserID: m.UserID,
		}
	case *tg.ChatParticipantAdmin:
		return &tg.ChannelParticipantAdmin{
			UserID:    m.UserID,
			InviterID: m.InviterID,
			Date:      m.Date,
		}
	case *tg.ChatParticipant:
		return &tg.ChannelParticipant{
			UserID: m.UserID,
			Date:   m.Date,
		}
	default:
		return &tg.ChannelParticipant{}
	}
}

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
		var err error
		inputPeer, err = ResolveInputPeerByID(ctx, raw, p, chatID)
		if err != nil {
			return nil, err
		}
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
			if tg.IsUserNotParticipant(err) {
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
		return nil, errors.ErrNotChat
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
