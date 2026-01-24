package storage

import (
	"log"
	"sync"
	"time"

	"github.com/AnimeKaizoku/cacher"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// PeerStorage provides a two-tier cache system for Telegram peers (users, chats, channels).
// It maintains an in-memory hot cache backed by a GORM database for persistence.
// Peers are automatically stored when encountered in updates and can be
// retrieved via GetInputPeerByID() for API calls.
type PeerStorage struct {
	peerCache  *cacher.Cacher[int64, *Peer]
	peerLock   *sync.RWMutex
	inMemory   bool
	SqlSession *gorm.DB
}

// NewPeerStorage creates a new peer storage instance with two-tier caching.
//
// Parameters:
//   - dialector: The GORM dialector for database connection
//   - inMemory: If true, only in-memory caching is used (no database persistence)
//
// The storage maintains:
//   - Hot cache: In-memory cache using AnimeKaizoku/cacher for fast lookups
//   - Cold storage: GORM database for persistence across restarts
//   - Auto-save: Peers from updates are automatically cached
//
// Returns:
//   - A new PeerStorage instance ready for use
//
// Example:
//
//	// In-memory storage (no database)
//	peerStorage := storage.NewPeerStorage(nil, true)
//
//	// SQLite with in-memory cache
//	dialector := sqlite.Open("sqlite.db")
//	peerStorage := storage.NewPeerStorage(dialector, false)
func NewPeerStorage(dialector gorm.Dialector, inMemory bool) *PeerStorage {
	p := PeerStorage{
		inMemory: inMemory,
		peerLock: new(sync.RWMutex),
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
		db, err := gorm.Open(dialector, &gorm.Config{
			SkipDefaultTransaction: true,
			Logger:                 logger.Default.LogMode(logger.Silent),
		})
		if err != nil {
			log.Panicln(err)
		}
		p.SqlSession = db
		dB, _ := db.DB()
		dB.SetMaxOpenConns(100)
		_ = p.SqlSession.AutoMigrate(&Session{}, &Peer{}, &ConvState{})
	}
	p.peerCache = cacher.NewCacher[int64, *Peer](opts)
	return &p
}
