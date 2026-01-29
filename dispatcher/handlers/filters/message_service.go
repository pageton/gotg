package filters

import (
	"github.com/gotd/td/tg"
	"github.com/pageton/gotg/types"
)

// Service returns true if the Message is a service message.
func (*messageFilters) Service(m *types.Message) bool {
	return m.IsService
}

// VideoChatStarted returns true if a video chat was started with this message.
func (*messageFilters) VideoChatStarted(m *types.Message) bool {
	if !m.IsService {
		return false
	}
	_, ok := m.Action.(*tg.MessageActionGroupCallScheduled)
	return ok
}

// VideoChatEnded returns true if a video chat ended with this message.
func (*messageFilters) VideoChatEnded(m *types.Message) bool {
	if !m.IsService {
		return false
	}
	action, ok := m.Action.(*tg.MessageActionGroupCall)
	return ok && action.Duration > 0
}

// VideoChatMembersInvited returns true if members were invited to a video chat.
func (*messageFilters) VideoChatMembersInvited(m *types.Message) bool {
	if !m.IsService {
		return false
	}
	_, ok := m.Action.(*tg.MessageActionInviteToGroupCall)
	return ok
}

// GameHighScore returns true if a game high score was achieved with this message.
func (*messageFilters) GameHighScore(m *types.Message) bool {
	if !m.IsService {
		return false
	}
	_, ok := m.Action.(*tg.MessageActionGameScore)
	return ok
}

func (*messageFilters) Giveaway(m *types.Message) bool {
	if m.IsService {
		_, ok := m.Action.(*tg.MessageActionGiveawayLaunch)
		return ok
	}
	return false
}

func (*messageFilters) GiveawayWinners(m *types.Message) bool {
	if m.IsService {
		_, ok := m.Action.(*tg.MessageActionGiveawayResults)
		return ok
	}
	return false
}

func (*messageFilters) GiftCode(m *types.Message) bool {
	if m.IsService {
		_, ok := m.Action.(*tg.MessageActionGiftCode)
		return ok
	}
	return false
}

func (*messageFilters) Gift(m *types.Message) bool {
	if m.IsService {
		_, ok := m.Action.(*tg.MessageActionGiftPremium)
		return ok
	}
	return false
}

func (*messageFilters) UsersShared(m *types.Message) bool {
	if m.IsService {
		_, ok := m.Action.(*tg.MessageActionRequestedPeer)
		return ok
	}
	return false
}

func (*messageFilters) ChatShared(m *types.Message) bool {
	if m.IsService {
		_, ok := m.Action.(*tg.MessageActionRequestedPeer)
		return ok
	}
	return false
}

func (*messageFilters) Forum(m *types.Message) bool {
	if !m.IsService {
		return false
	}
	_, ok := m.Action.(*tg.MessageActionTopicCreate)
	return ok
}

func (*messageFilters) NewChatMembers(m *types.Message) bool {
	if !m.IsService {
		return false
	}
	_, ok := m.Action.(*tg.MessageActionChatAddUser)
	return ok
}

func (*messageFilters) LeftChatMember(m *types.Message) bool {
	if !m.IsService {
		return false
	}
	_, ok := m.Action.(*tg.MessageActionChatDeleteUser)
	return ok
}

func (*messageFilters) NewChatTitle(m *types.Message) bool {
	if !m.IsService {
		return false
	}
	_, ok := m.Action.(*tg.MessageActionChatEditTitle)
	return ok
}

func (*messageFilters) NewChatPhoto(m *types.Message) bool {
	if !m.IsService {
		return false
	}
	_, ok := m.Action.(*tg.MessageActionChatEditPhoto)
	return ok
}

func (*messageFilters) DeleteChatPhoto(m *types.Message) bool {
	if !m.IsService {
		return false
	}
	_, ok := m.Action.(*tg.MessageActionChatDeletePhoto)
	return ok
}

func (*messageFilters) GroupChatCreated(m *types.Message) bool {
	if !m.IsService {
		return false
	}
	_, ok := m.Action.(*tg.MessageActionChatCreate)
	return ok
}

func (*messageFilters) SupergroupChatCreated(m *types.Message) bool {
	if !m.IsService {
		return false
	}
	_, ok := m.Action.(*tg.MessageActionChannelCreate)
	return ok
}

func (*messageFilters) ChannelChatCreated(m *types.Message) bool {
	if !m.IsService {
		return false
	}
	_, ok := m.Action.(*tg.MessageActionChannelCreate)
	return ok
}

func (*messageFilters) MigrateToChatID(m *types.Message) bool {
	if !m.IsService {
		return false
	}
	_, ok := m.Action.(*tg.MessageActionChatMigrateTo)
	return ok
}

func (*messageFilters) MigrateFromChatID(m *types.Message) bool {
	if !m.IsService {
		return false
	}
	_, ok := m.Action.(*tg.MessageActionChannelMigrateFrom)
	return ok
}

func (*messageFilters) PinnedMessage(m *types.Message) bool {
	return m.Pinned
}
