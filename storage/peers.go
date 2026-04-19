package storage

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gotd/td/constant"
	"github.com/gotd/td/telegram/query/dialogs"
	"github.com/gotd/td/tg"
)

// Username stores a Telegram username with its metadata (active, editable).
// Mirrors tg.Username but is JSON-safe for persistent storage adapters.
type Username struct {
	Username string
	Active   bool
	Editable bool
}

type Peer struct {
	ID          int64
	AccessHash  int64
	Type        int
	Username    string
	Usernames   Usernames
	PhoneNumber string
	IsBot       bool
	PhotoID     int64
	Language    string
	LastUpdated int64
}
type EntityType int

func (e EntityType) GetInt() int {
	return int(e)
}

const (
	DefaultUsername   = ""
	DefaultAccessHash = 0
	DefaultPhone      = ""
)

const (
	_ EntityType = iota
	TypeUser
	TypeChat
	TypeChannel
)

func (p *PeerStorage) AddPeer(iD, accessHash int64, peerType EntityType, userName string) {
	var usernames Usernames
	if userName != "" {
		usernames = Usernames{{Username: userName}}
	}
	p.AddPeerWithUsernames(iD, accessHash, peerType, userName, usernames, DefaultPhone, false, 0)
}

// AddPeerWithUsernames adds a peer with its primary username and all usernames (active + collectible).
func (p *PeerStorage) AddPeerWithUsernames(iD, accessHash int64, peerType EntityType, userName string, usernames Usernames, phoneNumber string, isBot bool, photoID int64) {
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

	p.peerLock.Lock()
	// Check if peer already exists in cache
	existingPeer, exists := p.peerCache.Get(iD)
	if exists && existingPeer != nil {
		// Don't overwrite a valid access_hash with zero (e.g. min peer).
		if accessHash == 0 && existingPeer.AccessHash != 0 {
			accessHash = existingPeer.AccessHash
		}
		// Update existing peer while preserving fields like Language
		existingPeer.AccessHash = accessHash
		existingPeer.Type = peerType.GetInt()
		existingPeer.Username = userName
		existingPeer.Usernames = usernames
		if phoneNumber != "" {
			existingPeer.PhoneNumber = phoneNumber
		}
		existingPeer.IsBot = isBot
		if photoID != 0 {
			existingPeer.PhotoID = photoID
		}
		existingPeer.LastUpdated = time.Now().Unix()
		peer = existingPeer
	} else {
		if !p.inMemory {
			if dbPeer, err := p.db.GetPeerByID(iD); err == nil && dbPeer != nil {
				// Don't overwrite a valid access_hash with zero (e.g. min peer).
				if accessHash == 0 && dbPeer.AccessHash != 0 {
					accessHash = dbPeer.AccessHash
				}
				dbPeer.AccessHash = accessHash
				dbPeer.Type = peerType.GetInt()
				dbPeer.Username = userName
				dbPeer.Usernames = usernames
				if phoneNumber != "" {
					dbPeer.PhoneNumber = phoneNumber
				}
				dbPeer.IsBot = isBot
				if photoID != 0 {
					dbPeer.PhotoID = photoID
				}
				dbPeer.LastUpdated = time.Now().Unix()
				peer = dbPeer
			} else {
				peer = &Peer{ID: iD, AccessHash: accessHash, Type: peerType.GetInt(), Username: userName, Usernames: usernames, PhoneNumber: phoneNumber, IsBot: isBot, PhotoID: photoID, LastUpdated: time.Now().Unix()}
			}
		} else {
			peer = &Peer{ID: iD, AccessHash: accessHash, Type: peerType.GetInt(), Username: userName, Usernames: usernames, PhoneNumber: phoneNumber, IsBot: isBot, PhotoID: photoID, LastUpdated: time.Now().Unix()}
		}
	}

	p.peerCache.Set(iD, peer)
	p.updateUsernameIndex(iD, peer)
	p.updatePhoneIndex(iD, peer)
	p.peerLock.Unlock()

	if p.inMemory {
		return
	}
	p.writeCh <- peer
}

// updateUsernameIndex removes old username mappings and adds new ones.
// Caller must hold peerLock.
func (p *PeerStorage) updateUsernameIndex(peerID int64, peer *Peer) {
	// Remove old usernames that belong to this peer.
	for u, id := range p.usernameIndex {
		if id == peerID {
			delete(p.usernameIndex, u)
		}
	}
	// Add current usernames.
	if peer.Username != "" {
		p.usernameIndex[peer.Username] = peerID
	}
	for _, u := range peer.Usernames {
		if u.Username != "" {
			p.usernameIndex[u.Username] = peerID
		}
	}
}

// updatePhoneIndex removes old phone mapping and adds the new one.
// Caller must hold peerLock.
func (p *PeerStorage) updatePhoneIndex(peerID int64, peer *Peer) {
	// Remove old phone that belongs to this peer.
	for ph, id := range p.phoneIndex {
		if id == peerID {
			delete(p.phoneIndex, ph)
		}
	}
	if peer.PhoneNumber != "" {
		p.phoneIndex[peer.PhoneNumber] = peerID
	}
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
// Strips leading @ and normalizes to lowercase before lookup.
// Uses username index for O(1) lookup; stale usernames are automatically cleaned up.
// On cache+DB miss, falls back to contacts.ResolveUsername RPC if a resolver is set.
// Returns nil if the peer is not found.
func (p *PeerStorage) GetPeerByUsername(username string) *Peer {
	username = strings.ToLower(strings.TrimPrefix(username, "@"))

	// 1. Check username index for O(1) lookup.
	p.peerLock.RLock()
	peerID, ok := p.usernameIndex[username]
	p.peerLock.RUnlock()
	if ok {
		if peer, ok := p.peerCache.Get(peerID); ok && peer != nil {
			return peer
		}
	}

	// 2. Check database.
	if !p.inMemory {
		peer, err := p.db.GetPeerByUsername(username)
		if err != nil {
			log.Printf("peers: failed to get peer by username %s: %v", username, err)
		} else if peer != nil {
			p.peerLock.Lock()
			p.peerCache.Set(peer.ID, peer)
			p.updateUsernameIndex(peer.ID, peer)
			p.updatePhoneIndex(peer.ID, peer)
			p.peerLock.Unlock()
			return peer
		}
	}

	// 3. RPC fallback via contacts.ResolveUsername.
	if p.resolveUsernameFn != nil {
		users, chats, err := p.resolveUsernameFn(context.Background(), username)
		if err != nil {
			log.Printf("peers: RPC resolve username %s failed: %v", username, err)
			return nil
		}
		saveResolvedPeers(p, users, chats)
		// Retry index lookup after RPC response was saved.
		p.peerLock.RLock()
		peerID, ok = p.usernameIndex[username]
		p.peerLock.RUnlock()
		if ok {
			if peer, ok := p.peerCache.Get(peerID); ok && peer != nil {
				return peer
			}
		}
	}

	return nil
}

// GetPeerByPhoneNumber finds a peer by phone number.
// Uses phone index for O(1) lookup. On cache+DB miss, falls back to
// contacts.ResolvePhone RPC if a resolver is set.
// Returns nil if the peer is not found.
func (p *PeerStorage) GetPeerByPhoneNumber(phone string) *Peer {
	// 1. Check phone index for O(1) lookup.
	p.peerLock.RLock()
	peerID, ok := p.phoneIndex[phone]
	p.peerLock.RUnlock()
	if ok {
		if peer, ok := p.peerCache.Get(peerID); ok && peer != nil {
			return peer
		}
	}

	// 2. Check database.
	if !p.inMemory {
		peer, err := p.db.GetPeerByPhoneNumber(phone)
		if err != nil {
			log.Printf("peers: failed to get peer by phone %s: %v", phone, err)
		} else if peer != nil {
			p.peerLock.Lock()
			p.peerCache.Set(peer.ID, peer)
			p.updateUsernameIndex(peer.ID, peer)
			p.updatePhoneIndex(peer.ID, peer)
			p.peerLock.Unlock()
			return peer
		}
	}

	// 3. RPC fallback via contacts.ResolvePhone.
	if p.resolvePhoneFn != nil {
		users, chats, err := p.resolvePhoneFn(context.Background(), phone)
		if err != nil {
			log.Printf("peers: RPC resolve phone %s failed: %v", phone, err)
			return nil
		}
		saveResolvedPeers(p, users, chats)
		// Retry index lookup after RPC response was saved.
		p.peerLock.RLock()
		peerID, ok = p.phoneIndex[phone]
		p.peerLock.RUnlock()
		if ok {
			if peer, ok := p.peerCache.Get(peerID); ok && peer != nil {
				return peer
			}
		}
	}

	return nil
}

// ExtractPhotoID extracts the photo_id from a UserProfilePhotoClass.
// Returns 0 if the photo is nil or not a *tg.UserProfilePhoto.
func ExtractPhotoID(photo tg.UserProfilePhotoClass) int64 {
	if photo == nil {
		return 0
	}
	p, ok := photo.(*tg.UserProfilePhoto)
	if !ok {
		return 0
	}
	return p.PhotoID
}

// ExtractChatPhotoID extracts the photo_id from a ChatPhotoClass.
// Returns 0 if the photo is nil or not a *tg.ChatPhoto.
func ExtractChatPhotoID(photo tg.ChatPhotoClass) int64 {
	if photo == nil {
		return 0
	}
	p, ok := photo.(*tg.ChatPhoto)
	if !ok {
		return 0
	}
	return p.PhotoID
}

// saveResolvedPeers saves users and chats from a contacts.ResolveUsername/ResolvePhone response
// into peer storage, with min-peer filtering and lowercase normalization.
func saveResolvedPeers(p *PeerStorage, users []tg.UserClass, chats []tg.ChatClass) {
	for _, user := range users {
		u, ok := user.(*tg.User)
		if !ok {
			continue
		}
		if u.Min {
			continue
		}
		p.AddPeerWithUsernames(u.ID, u.AccessHash, TypeUser, strings.ToLower(u.Username), ConvertUsernames(u.Usernames), u.Phone, u.Bot, ExtractPhotoID(u.Photo))
	}
	for _, chat := range chats {
		channel, ok := chat.(*tg.Channel)
		if ok {
			if !channel.Min {
				p.AddPeerWithUsernames(channel.ID, channel.AccessHash, TypeChannel, strings.ToLower(channel.Username), ConvertUsernames(channel.Usernames), DefaultPhone, false, 0)
			}
			continue
		}
		c, ok := chat.(*tg.Chat)
		if !ok {
			continue
		}
		p.AddPeerWithUsernames(c.ID, DefaultAccessHash, TypeChat, DefaultUsername, nil, DefaultPhone, false, ExtractChatPhotoID(c.Photo))
	}
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
	p.peerLock.Lock()
	p.peerCache.Set(id, peer)
	p.updateUsernameIndex(id, peer)
	p.updatePhoneIndex(id, peer)
	p.peerLock.Unlock()
	return peer
}

func (p *PeerStorage) SetPeerLanguage(userID int64, lang string) {
	peer := p.GetPeerByID(userID)
	p.peerLock.Lock()
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
	p.peerLock.Unlock()

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

// ConvertUsernames converts []tg.Username to Usernames with lowercase normalization,
// preserving Active and Editable metadata.
func ConvertUsernames(usernames []tg.Username) Usernames {
	if len(usernames) == 0 {
		return nil
	}
	result := make(Usernames, len(usernames))
	for i, u := range usernames {
		result[i] = Username{
			Username: strings.ToLower(u.Username),
			Active:   u.Active,
			Editable: u.Editable,
		}
	}
	return result
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
			if channel.Min {
				continue
			}
			peerStorage.AddPeerWithUsernames(cid, channel.AccessHash, TypeChannel, strings.ToLower(channel.Username), ConvertUsernames(channel.Usernames), DefaultPhone, false, 0)
		}
		for uid, user := range e.Entities.Users() {
			if user.Min {
				continue
			}
			peerStorage.AddPeerWithUsernames(uid, user.AccessHash, TypeUser, strings.ToLower(user.Username), ConvertUsernames(user.Usernames), user.Phone, user.Bot, ExtractPhotoID(user.Photo))
		}
		for gid, chat := range e.Entities.Chats() {
			peerStorage.AddPeerWithUsernames(gid, DefaultAccessHash, TypeChat, DefaultUsername, nil, DefaultPhone, false, ExtractChatPhotoID(chat.Photo))
		}
		return nil
	})
}
