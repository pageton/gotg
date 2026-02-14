package storage

import (
	"errors"

	"gorm.io/gorm"
)

type gormAdapterCompat struct {
	db *gorm.DB
}

func (g *gormAdapterCompat) DB() *gorm.DB { return g.db }

func (g *gormAdapterCompat) GetSession(version int) (*Session, error) {
	s := &Session{Version: version}
	if err := g.db.Model(&Session{}).First(s, "version = ?", version).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return s, nil
}

func (g *gormAdapterCompat) UpdateSession(s *Session) error {
	tx := g.db.Begin()
	if err := tx.Save(s).Error; err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

func (g *gormAdapterCompat) SavePeer(p *Peer) error {
	return g.db.Save(p).Error
}

func (g *gormAdapterCompat) GetPeerByID(id int64) (*Peer, error) {
	peer := &Peer{}
	if err := g.db.Where("id = ?", id).First(peer).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return peer, nil
}

func (g *gormAdapterCompat) GetPeerByUsername(username string) (*Peer, error) {
	peer := &Peer{}
	if err := g.db.Where("username = ?", username).First(peer).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return peer, nil
}

func (g *gormAdapterCompat) SaveConvState(state *ConvState) error {
	return g.db.Save(state).Error
}

func (g *gormAdapterCompat) LoadConvState(key string) (*ConvState, error) {
	var state ConvState
	if err := g.db.Where("key = ?", key).First(&state).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &state, nil
}

func (g *gormAdapterCompat) DeleteConvState(key string) error {
	return g.db.Delete(&ConvState{Key: key}).Error
}

func (g *gormAdapterCompat) ListConvStates() ([]ConvState, error) {
	var states []ConvState
	if err := g.db.Find(&states).Error; err != nil {
		return nil, err
	}
	return states, nil
}

func (g *gormAdapterCompat) AutoMigrate() error {
	return g.db.AutoMigrate(&Session{}, &Peer{}, &ConvState{})
}

func (g *gormAdapterCompat) Close() error {
	sqlDB, err := g.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
