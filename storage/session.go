package storage

import "log"

type Session struct {
	Version int `gorm:"primary_key"`
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
