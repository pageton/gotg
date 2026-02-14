// Package gormdb provides a GORM-based Backend implementation.
// It works with any GORM dialector: PostgreSQL, MySQL, SQLite (modernc or mattn).
//
// Usage:
//
//	// PostgreSQL
//	b, _ := gormdb.New(postgres.Open(dsn))
//
//	// MySQL
//	b, _ := gormdb.New(mysql.Open(dsn))
//
//	// SQLite (modernc, pure Go)
//	b, _ := gormdb.New(sqlite.Open("bot.db"))
//
//	// SQLite (mattn, CGO)
//	b, _ := gormdb.New(sqlite.Open("bot.db"))
package gormdb

import (
	"errors"
	"time"

	"github.com/pageton/gotg/storage"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type GormAdapter struct {
	db *gorm.DB
}

// Option configures a GormBackend.
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

// New creates a new GormBackend from a GORM dialector.
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

	return &GormAdapter{db: db}, nil
}

// DB returns the underlying *gorm.DB for advanced usage.
func (g *GormAdapter) DB() *gorm.DB {
	return g.db
}

func (g *GormAdapter) GetSession(version int) (*storage.Session, error) {
	s := &storage.Session{Version: version}
	if err := g.db.Model(&storage.Session{}).First(s, "version = ?", version).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return s, nil
}

func (g *GormAdapter) UpdateSession(s *storage.Session) error {
	tx := g.db.Begin()
	if err := tx.Save(s).Error; err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

func (g *GormAdapter) SavePeer(p *storage.Peer) error {
	return g.db.Save(p).Error
}

func (g *GormAdapter) GetPeerByID(id int64) (*storage.Peer, error) {
	peer := &storage.Peer{}
	if err := g.db.Where("id = ?", id).First(peer).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return peer, nil
}

func (g *GormAdapter) GetPeerByUsername(username string) (*storage.Peer, error) {
	peer := &storage.Peer{}
	if err := g.db.Where("username = ?", username).First(peer).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return peer, nil
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

func (g *GormAdapter) AutoMigrate() error {
	return g.db.AutoMigrate(&storage.Session{}, &storage.Peer{}, &storage.ConvState{})
}

func (g *GormAdapter) Close() error {
	sqlDB, err := g.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
