package storage

import "log"

// Session stores the MTProto session data (auth key and related state).
// The Data field contains the raw 256-byte auth key.
// Anyone with database access can use this key to impersonate the Telegram session.
// Enable session encryption via PeerStorage.SetEncryptor to protect Data at rest.
type Session struct {
	Version int
	Data    []byte
}

const LatestVersion = 1

func (p *PeerStorage) UpdateSession(session *Session) {
	if p.db == nil {
		return
	}
	toStore := session
	if p.encryptor != nil && len(session.Data) > 0 {
		encrypted, err := p.encryptor.Encrypt(session.Data)
		if err != nil {
			log.Printf("session: encrypt failed: %v", err)
			return
		}
		toStore = &Session{Version: session.Version, Data: encrypted}
	}
	if err := p.db.UpdateSession(toStore); err != nil {
		log.Printf("session: failed to update: %v", err)
	}
}

func (p *PeerStorage) GetSession() *Session {
	if p.db == nil {
		return &Session{Version: LatestVersion}
	}
	s, err := p.db.GetSession(LatestVersion)
	if err != nil {
		log.Printf("session: failed to get: %v", err)
		return &Session{Version: LatestVersion}
	}
	if s == nil {
		return &Session{Version: LatestVersion}
	}
	if p.encryptor != nil && len(s.Data) > 0 {
		decrypted, err := p.encryptor.Decrypt(s.Data)
		if err != nil {
			log.Printf("session: decrypt failed: %v", err)
			return &Session{Version: LatestVersion}
		}
		s.Data = decrypted
	}
	return s
}
