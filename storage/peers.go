package storage

import (
	"context"
	"fmt"
	"log"

	"github.com/gotd/td/constant"
	"github.com/gotd/td/telegram/query/dialogs"
	"github.com/gotd/td/tg"
)

type Peer struct {
	ID         int64 `gorm:"primary_key"`
	AccessHash int64
	Type       int
	Username   string `gorm:"index"`
	Language   string
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

	var peer *Peer

	// Check if peer already exists in cache
	existingPeer, exists := p.peerCache.Get(iD)
	if exists && existingPeer != nil {
		// Update existing peer while preserving fields like Language
		existingPeer.AccessHash = accessHash
		existingPeer.Type = peerType.GetInt()
		existingPeer.Username = userName
		peer = existingPeer
	} else {
		if !p.inMemory {
			if dbPeer, err := p.db.GetPeerByID(iD); err == nil && dbPeer != nil {
				dbPeer.AccessHash = accessHash
				dbPeer.Type = peerType.GetInt()
				dbPeer.Username = userName
				peer = dbPeer
			} else {
				peer = &Peer{ID: iD, AccessHash: accessHash, Type: peerType.GetInt(), Username: userName}
			}
		} else {
			peer = &Peer{ID: iD, AccessHash: accessHash, Type: peerType.GetInt(), Username: userName}
		}
	}

	p.peerCache.Set(iD, peer)
	if p.inMemory {
		return
	}
	p.writeCh <- peer
}

func (p *PeerStorage) startWriter() {
	defer func() {
		if p.writerDone != nil {
			close(p.writerDone)
		}
	}()
	for peer := range p.writeCh {
		p.peerLock.Lock()
		if err := p.db.SavePeer(peer); err != nil {
			log.Printf("peers: failed to save peer %d to database: %v", peer.ID, err)
		}
		p.peerLock.Unlock()
	}
}

func (p *Peer) GetID() int64 {
	switch EntityType(p.Type) {
	case TypeChat, TypeChannel:
		tdlibID := constant.TDLibPeerID(p.ID)
		return tdlibID.ToPlain()
	default:
		return p.ID
	}
}

// GetPeerByID finds the provided id in the peer storage and return it if found.
func (p *PeerStorage) GetPeerByID(peerID int64) *Peer {
	peer, ok := p.peerCache.Get(peerID)
	if p.inMemory {
		if !ok {
			return nil
		}
	} else {
		if !ok {
			peer = p.cachePeers(peerID)
			// Return nil if peer doesn't exist in DB (ID is 0)
			if peer != nil && peer.ID == 0 {
				return nil
			}
		}
	}
	return peer
}

// GetPeerByUsername finds the provided username in the peer storage and return it if found.
// Returns nil if the peer is not found.
func (p *PeerStorage) GetPeerByUsername(username string) *Peer {
	if p.inMemory {
		for _, peer := range p.peerCache.GetAll() {
			if peer.Username == username {
				return peer
			}
		}
	} else {
		peer, err := p.db.GetPeerByUsername(username)
		if err != nil {
			log.Printf("peers: failed to get peer by username %s: %v", username, err)
			return nil
		}
		if peer != nil {
			return peer
		}
	}
	return nil
}

// GetInputPeerByID finds the provided id in the peer storage and return its tg.InputPeerClass if found.
func (p *PeerStorage) GetInputPeerByID(iD int64) tg.InputPeerClass {
	return getInputPeerFromStoragePeer(p.GetPeerByID(iD))
}

// GetInputPeerByUsername finds the provided username in the peer storage and return its tg.InputPeerClass if found.
func (p *PeerStorage) GetInputPeerByUsername(userName string) tg.InputPeerClass {
	return getInputPeerFromStoragePeer(p.GetPeerByUsername(userName))
}

func (p *PeerStorage) cachePeers(id int64) *Peer {
	peer, err := p.db.GetPeerByID(id)
	if err != nil {
		log.Printf("peers: failed to get peer %d: %v", id, err)
		return nil
	}
	if peer == nil {
		return nil
	}
	p.peerCache.Set(id, peer)
	return peer
}

func (p *PeerStorage) SetPeerLanguage(userID int64, lang string) {
	peer := p.GetPeerByID(userID)
	if peer == nil {
		peer = &Peer{
			ID:       userID,
			Language: lang,
			Type:     int(TypeUser),
		}
	} else {
		peer.Language = lang
	}
	p.peerCache.Set(userID, peer)

	if !p.inMemory {
		p.writeCh <- peer
	}
}

func getInputPeerFromStoragePeer(peer *Peer) tg.InputPeerClass {
	if peer == nil {
		return &tg.InputPeerEmpty{}
	}
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
			ChannelID:  ID.ToPlain(),
			AccessHash: peer.AccessHash,
		}
	default:
		return &tg.InputPeerEmpty{}
	}
}

// AddPeersFromDialogs iterates all dialogs via the Telegram API and adds
// every encountered user, chat and channel to peerStorage.
// It returns any error from the underlying RPC pagination.
func AddPeersFromDialogs(ctx context.Context, raw *tg.Client, peerStorage *PeerStorage) error {
	return dialogs.NewQueryBuilder(raw).GetDialogs().BatchSize(100).ForEach(ctx, func(ctx context.Context, e dialogs.Elem) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
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
