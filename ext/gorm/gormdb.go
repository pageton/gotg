// Package gormdb provides a GORM-based storage adapter.
// It works with any GORM dialector: PostgreSQL, MySQL, SQLite (modernc or mattn).
//
// Quick start with SQLite:
//
//	client, _ := gotg.NewClient(appID, hash, gotg.AsBot(token), &gotg.ClientOpts{
//	    Session: gormdb.SqliteSession("bot.db"),
//	})
//
// Advanced usage with any dialector:
//
//	adapter, _ := gormdb.New(postgres.Open(dsn))
//	client, _ := gotg.NewClient(appID, hash, gotg.AsBot(token), &gotg.ClientOpts{
//	    Session: session.WithAdapter(adapter),
//	})
package gormdb

import (
	"errors"
	"time"

	"github.com/bytedance/sonic"
	"github.com/pageton/gotg/session"
	"github.com/pageton/gotg/storage"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type GormAdapter struct {
	db *gorm.DB
}

// Option configures a GormAdapter.
type Option func(*config)

type config struct {
	gormConfig      *gorm.Config
	maxOpenConns    int
	maxIdleConns    int
	connMaxLifetime time.Duration
	connMaxIdleTime time.Duration
	skipDefaultTx   bool
	silentLogger    bool
}

func defaults() *config {
	return &config{
		maxOpenConns:    100,
		maxIdleConns:    10,
		connMaxLifetime: 30 * time.Minute,
		connMaxIdleTime: 5 * time.Minute,
		skipDefaultTx:   true,
		silentLogger:    true,
	}
}

// WithGormConfig overrides the default GORM config.
func WithGormConfig(cfg *gorm.Config) Option {
	return func(c *config) { c.gormConfig = cfg }
}

// WithMaxOpenConns sets the maximum number of open connections.
func WithMaxOpenConns(n int) Option {
	return func(c *config) { c.maxOpenConns = n }
}

// WithMaxIdleConns sets the maximum number of idle connections.
func WithMaxIdleConns(n int) Option {
	return func(c *config) { c.maxIdleConns = n }
}

// WithConnMaxLifetime sets the maximum connection lifetime.
func WithConnMaxLifetime(d time.Duration) Option {
	return func(c *config) { c.connMaxLifetime = d }
}

// WithConnMaxIdleTime sets the maximum idle connection time.
func WithConnMaxIdleTime(d time.Duration) Option {
	return func(c *config) { c.connMaxIdleTime = d }
}

// New creates a new GormAdapter from a GORM dialector.
// It also runs AutoMigrate to ensure all required tables exist.
func New(dialector gorm.Dialector, opts ...Option) (*GormAdapter, error) {
	cfg := defaults()
	for _, o := range opts {
		o(cfg)
	}

	gormCfg := cfg.gormConfig
	if gormCfg == nil {
		gormCfg = &gorm.Config{
			SkipDefaultTransaction: cfg.skipDefaultTx,
		}
		if cfg.silentLogger {
			gormCfg.Logger = logger.Default.LogMode(logger.Silent)
		}
	}

	db, err := gorm.Open(dialector, gormCfg)
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(cfg.maxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.maxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.connMaxLifetime)
	sqlDB.SetConnMaxIdleTime(cfg.connMaxIdleTime)

	adapter := &GormAdapter{db: db}
	if err := adapter.AutoMigrate(); err != nil {
		return nil, err
	}
	return adapter, nil
}

// DB returns the underlying *gorm.DB for advanced usage.
func (g *GormAdapter) DB() *gorm.DB {
	return g.db
}

func (g *GormAdapter) GetSession(version int) (*storage.Session, error) {
	s := &gormSession{Version: version}
	if err := g.db.Model(&gormSession{}).First(s, "version = ?", version).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &storage.Session{Version: s.Version, Data: s.Data}, nil
}

func (g *GormAdapter) UpdateSession(s *storage.Session) error {
	tx := g.db.Begin()
	if err := tx.Save(&gormSession{Version: s.Version, Data: s.Data}).Error; err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

func (g *GormAdapter) SavePeer(p *storage.Peer) error {
	return g.db.Save(newGormPeer(p)).Error
}

func (g *GormAdapter) GetPeerByID(id int64) (*storage.Peer, error) {
	peer := &gormPeer{}
	if err := g.db.Where("id = ?", id).First(peer).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return peer.toStoragePeer(), nil
}

func (g *GormAdapter) GetPeerByUsername(username string) (*storage.Peer, error) {
	peer := &gormPeer{}
	if err := g.db.Where("username = ?", username).First(peer).Error; err == nil {
		return peer.toStoragePeer(), nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	// Search in usernames JSON column using LIKE — portable across all SQL dialects.
	// Usernames are lowercase alphanumeric + underscores, safe to embed in a LIKE pattern.
	peer = &gormPeer{}
	if err := g.db.Where("usernames LIKE ?", "%\""+username+"\"%").First(peer).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return peer.toStoragePeer(), nil
}

func (g *GormAdapter) GetPeerByPhoneNumber(phone string) (*storage.Peer, error) {
	peer := &gormPeer{}
	if err := g.db.Where("phone_number = ?", phone).First(peer).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return peer.toStoragePeer(), nil
}

func (g *GormAdapter) SaveConvState(state *storage.ConvState) error {
	return g.db.Save(state).Error
}

func (g *GormAdapter) LoadConvState(key string) (*storage.ConvState, error) {
	var state storage.ConvState
	if err := g.db.Where("key = ?", key).First(&state).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &state, nil
}

func (g *GormAdapter) DeleteConvState(key string) error {
	return g.db.Delete(&storage.ConvState{Key: key}).Error
}

func (g *GormAdapter) ListConvStates() ([]storage.ConvState, error) {
	var states []storage.ConvState
	if err := g.db.Find(&states).Error; err != nil {
		return nil, err
	}
	return states, nil
}

type gormPeer struct {
	ID          int64  `gorm:"primaryKey"`
	AccessHash  int64
	Type        int
	Username    string
	Usernames   string `gorm:"type:text"`
	PhoneNumber string
	IsBot       bool
	PhotoID     int64
	Language    string
	LastUpdated int64
}

type gormSession struct {
	Version int `gorm:"primaryKey"`
	Data    []byte
}

func (gormSession) TableName() string { return "sessions" }

func (gormPeer) TableName() string { return "peers" }

func newGormPeer(p *storage.Peer) *gormPeer {
	if p == nil {
		return nil
	}
	usernames, _ := sonic.Marshal(p.Usernames)
	return &gormPeer{
		ID:          p.ID,
		AccessHash:  p.AccessHash,
		Type:        p.Type,
		Username:    p.Username,
		Usernames:   string(usernames),
		PhoneNumber: p.PhoneNumber,
		IsBot:       p.IsBot,
		PhotoID:     p.PhotoID,
		Language:    p.Language,
		LastUpdated: p.LastUpdated,
	}
}

func (p *gormPeer) toStoragePeer() *storage.Peer {
	if p == nil {
		return nil
	}
	var usernames storage.Usernames
	if p.Usernames != "" {
		_ = sonic.Unmarshal([]byte(p.Usernames), &usernames)
	}
	return &storage.Peer{
		ID:          p.ID,
		AccessHash:  p.AccessHash,
		Type:        p.Type,
		Username:    p.Username,
		Usernames:   usernames,
		PhoneNumber: p.PhoneNumber,
		IsBot:       p.IsBot,
		PhotoID:     p.PhotoID,
		Language:    p.Language,
		LastUpdated: p.LastUpdated,
	}
}

func (g *GormAdapter) AutoMigrate() error {
	return g.db.AutoMigrate(&gormSession{}, &gormPeer{}, &storage.ConvState{})
}

func (g *GormAdapter) DeleteStalePeers(olderThan int64) (int64, error) {
	result := g.db.Where("last_updated > 0 AND last_updated < ?", olderThan).Delete(&gormPeer{})
	return result.RowsAffected, result.Error
}

func (g *GormAdapter) Close() error {
	sqlDB, err := g.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// SqliteSession creates a GORM adapter backed by SQLite and returns a session
// constructor ready to pass to gotg.NewClient. It also runs AutoMigrate.
//
// Usage:
//
//	client, err := gotg.NewClient(appID, hash, gotg.AsBot(token), &gotg.ClientOpts{
//	    Session: gormdb.SqliteSession("bot.db"),
//	})
func Session(dialector gorm.Dialector, opts ...Option) *session.AdapterSessionConstructor {
	adapter, err := New(dialector, opts...)
	if err != nil {
		panic("gormdb: failed to open dialector: " + err.Error())
	}
	return session.Adapter(adapter)
}

// SqliteSession creates a GORM adapter backed by SQLite and returns a session
// constructor ready to pass to gotg.NewClient. It also runs AutoMigrate.
// Deprecated: use Session(sqlite.Open(dsn), opts...) instead.
func SqliteSession(dsn string, opts ...Option) *session.AdapterSessionConstructor {
	adapter, err := New(sqlite.Open(dsn), opts...)
	if err != nil {
		panic("gormdb: failed to open SQLite: " + err.Error())
	}
	return session.Adapter(adapter)
}
