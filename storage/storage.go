package storage

import (
	"fmt"
	"sync"
	"time"

	"github.com/AnimeKaizoku/cacher"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// GormDB is an optional interface that GORM-based backends can implement
// to expose the underlying *gorm.DB for backward compatibility.
type GormDB interface {
	DB() *gorm.DB
}

// PeerStorage provides a two-tier cache system for Telegram peers (users, chats, channels).
// It maintains an in-memory hot cache backed by a pluggable Adapter for persistence.
type PeerStorage struct {
	peerCache  *cacher.Cacher[int64, *Peer]
	peerLock   *sync.RWMutex
	inMemory   bool
	db         Adapter
	SqlSession *gorm.DB // Non-nil only for GORM backends (backward compat).
	writeCh    chan *Peer
	writerDone chan struct{}
}

// NewPeerStorageWithAdapter creates PeerStorage with a pluggable Adapter.
func NewPeerStorageWithAdapter(db Adapter, inMemory bool) (*PeerStorage, error) {
	p := PeerStorage{
		inMemory: inMemory,
		peerLock: new(sync.RWMutex),
		db:       db,
	}

	var opts *cacher.NewCacherOpts
	if inMemory {
		opts = nil
	} else {
		opts = &cacher.NewCacherOpts{
			TimeToLive:    6 * time.Hour,
			CleanInterval: 24 * time.Hour,
			Revaluate:     true,
		}
		if gb, ok := db.(GormDB); ok {
			p.SqlSession = gb.DB()
		}
		if err := p.db.AutoMigrate(); err != nil {
			return nil, fmt.Errorf("storage: auto-migrate failed: %w", err)
		}
	}

	p.peerCache = cacher.NewCacher[int64, *Peer](opts)
	if !inMemory {
		p.writeCh = make(chan *Peer, 256)
		p.writerDone = make(chan struct{})
		go p.startWriter()
	}
	return &p, nil
}

// NewPeerStorage creates PeerStorage from a GORM dialector (backward-compatible API).
func NewPeerStorage(dialector gorm.Dialector, inMemory bool) (*PeerStorage, error) {
	if inMemory || dialector == nil {
		return newInMemoryPeerStorage(inMemory)
	}

	db, err := gorm.Open(dialector, &gorm.Config{
		SkipDefaultTransaction: true,
		Logger:                 logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("storage: open database: %w", err)
	}

	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)
	sqlDB.SetConnMaxIdleTime(5 * time.Minute)

	adapter := &gormAdapterCompat{db: db}
	return NewPeerStorageWithAdapter(adapter, false)
}

func newInMemoryPeerStorage(inMemory bool) (*PeerStorage, error) {
	return NewPeerStorageWithAdapter(&memoryAdapterCompat{
		sessions:   make(map[int]*Session),
		peers:      make(map[int64]*Peer),
		convStates: make(map[string]*ConvState),
	}, inMemory)
}

func (p *PeerStorage) GetAdapter() Adapter {
	return p.db
}

// Close closes the write channel and waits for the writer goroutine to drain.
// Safe to call on in-memory storage (no-op).
func (p *PeerStorage) Close() {
	if p.writeCh != nil {
		close(p.writeCh)
		<-p.writerDone
	}
}
