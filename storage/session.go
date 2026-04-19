package storage

import "log"

// Session stores the MTProto session data (auth key and related state).
// The Data field contains the raw 256-byte auth key.
// Anyone with database access can use this key to impersonate the Telegram session.
// Consider encrypting Data at the application layer before persisting.
type Session struct {
	Version int
	Data    []byte
}

const LatestVersion = 1

func (p *PeerStorage) UpdateSession(session *Session) {
	if p.db == nil {
		return
	}
	if err := p.db.UpdateSession(session); err != nil {
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
	return s
}
