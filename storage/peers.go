package storage

import (
	"context"
	"fmt"

	"github.com/gotd/td/constant"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/telegram/query/dialogs"
)

type Peer struct {
	ID         int64 `gorm:"primary_key"`
	AccessHash int64
	Type       int
	Username   string
}
type EntityType int

func (e EntityType) GetInt() int {
	return int(e)
}

const (
	DefaultUsername   = ""
	DefaultAccessHash = 0
)

const (
	_ EntityType = iota
	TypeUser
	TypeChat
	TypeChannel
)

func (p *PeerStorage) AddPeer(iD, accessHash int64, peerType EntityType, userName string) {
	var ID constant.TDLibPeerID
	switch EntityType(peerType) {
		case TypeUser:
			ID.User(iD)
		case TypeChat:
			ID.Chat(iD)
		case TypeChannel:
			ID.Channel(iD)
	}
	iD = int64(ID)
	peer := &Peer{ID: iD, AccessHash: accessHash, Type: peerType.GetInt(), Username: userName}
	p.peerCache.Set(iD, peer)
	if p.inMemory {
		return
	}
	go p.addPeerToDb(peer)
}

func (p *PeerStorage) addPeerToDb(peer *Peer) {
	tx := p.SqlSession.Begin()
	tx.Save(peer)
	p.peerLock.Lock()
	defer p.peerLock.Unlock()
	tx.Commit()
}

// GetPeerById finds the provided id in the peer storage and return it if found.
func (p *PeerStorage) GetPeerById(iD int64) *Peer {
	peer, ok := p.peerCache.Get(iD)
	if p.inMemory {
		if !ok {
			return &Peer{}
		}
	} else {
		if !ok {
			return p.cachePeers(iD)
		}
	}
	return peer
}

// GetPeerByUsername finds the provided username in the peer storage and return it if found.
func (p *PeerStorage) GetPeerByUsername(username string) *Peer {
	if p.inMemory {
		for _, peer := range p.peerCache.GetAll() {
			if peer.Username == username {
				return peer
			}
		}
	} else {
		peer := Peer{}
		p.SqlSession.Where("username = ?", username).Find(&peer)
		return &peer
	}
	return &Peer{}
}

// GetInputPeerById finds the provided id in the peer storage and return its tg.InputPeerClass if found.
func (p *PeerStorage) GetInputPeerById(iD int64) tg.InputPeerClass {
	return getInputPeerFromStoragePeer(p.GetPeerById(iD))
}

// GetInputPeerByUsername finds the provided username in the peer storage and return its tg.InputPeerClass if found.
func (p *PeerStorage) GetInputPeerByUsername(userName string) tg.InputPeerClass {
	return getInputPeerFromStoragePeer(p.GetPeerByUsername(userName))
}

func (p *PeerStorage) cachePeers(id int64) *Peer {
	var peer = Peer{}
	p.SqlSession.Where("id = ?", id).Find(&peer)
	p.peerCache.Set(id, &peer)
	return &peer
}

func getInputPeerFromStoragePeer(peer *Peer) tg.InputPeerClass {
	ID := constant.TDLibPeerID(peer.ID)
	warning := "DEPRECATION: Fetching PeerID from non-BotAPI IDs is deprecated — Please use Bot API-style IDs (%s<id>) Instead.\n"
	switch EntityType(peer.Type) {
	case TypeUser:
		return &tg.InputPeerUser{
			UserID:     peer.ID,
			AccessHash: peer.AccessHash,
		}
	case TypeChat:
		if !ID.IsChat() {
			fmt.Printf(warning, "-")
		}
		return &tg.InputPeerChat{
			ChatID: ID.ToPlain(),
		}
	case TypeChannel:
		if !ID.IsChannel() {
			fmt.Printf(warning, "-100")
		}
		return &tg.InputPeerChannel{
			ChannelID: ID.ToPlain(),
			AccessHash: peer.AccessHash,
		}
	default:
		return &tg.InputPeerEmpty{}
	}
}


func AddPeersFromDialogs(ctx context.Context, raw *tg.Client, peerStorage *PeerStorage) {
	_ = dialogs.NewQueryBuilder(raw).GetDialogs().ForEach(ctx, func(ctx context.Context, e dialogs.Elem) error {
    for cid, channel := range e.Entities.Channels() {
      peerStorage.AddPeer(cid, channel.AccessHash, TypeChannel, channel.Username)
    }
    for uid, user := range e.Entities.Users() {
      peerStorage.AddPeer(uid, user.AccessHash, TypeUser, user.Username)
    }
    for gid := range e.Entities.Chats() {
      peerStorage.AddPeer(gid, DefaultAccessHash, TypeChat, DefaultUsername)
    }
    return nil
  })
}
