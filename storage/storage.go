package storage

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/AnimeKaizoku/cacher"
	"github.com/gotd/td/tg"
)

// PeerStorage provides a two-tier cache system for Telegram peers (users, chats, channels).
// It maintains an in-memory hot cache backed by a pluggable Adapter for persistence.
type PeerStorage struct {
	peerCache         *cacher.Cacher[int64, *Peer]
	peerLock          *sync.RWMutex
	usernameIndex     map[string]int64 // username -> peer ID
	phoneIndex        map[string]int64 // phone -> peer ID
	inMemory          bool
	db                Adapter
	writeCh           chan *Peer
	writerDone        chan struct{}
	stopPrune         chan struct{}
	pruneDone         chan struct{}
	resolveUsernameFn func(ctx context.Context, username string) ([]tg.UserClass, []tg.ChatClass, error)
	resolvePhoneFn    func(ctx context.Context, phone string) ([]tg.UserClass, []tg.ChatClass, error)
}

const pruneInterval = 24 * time.Hour

// NewPeerStorageWithAdapter creates PeerStorage with a pluggable Adapter.
func NewPeerStorageWithAdapter(db Adapter, inMemory bool) (*PeerStorage, error) {
	p := PeerStorage{
		inMemory:      inMemory,
		peerLock:      new(sync.RWMutex),
		usernameIndex: make(map[string]int64),
		phoneIndex:    make(map[string]int64),
		db:            db,
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
		if err := p.db.AutoMigrate(); err != nil {
			return nil, fmt.Errorf("storage: auto-migrate failed: %w", err)
		}
	}

	p.peerCache = cacher.NewCacher[int64, *Peer](opts)
	if !inMemory {
		p.writeCh = make(chan *Peer, 256)
		p.writerDone = make(chan struct{})
		go p.startWriter()

		p.stopPrune = make(chan struct{})
		p.pruneDone = make(chan struct{})
		go p.startPruner()
	}
	return &p, nil
}

func (p *PeerStorage) GetAdapter() Adapter {
	return p.db
}

// SetResolver injects a callback that resolves a username via the Telegram API
// (contacts.ResolveUsername RPC). Called automatically by GetPeerByUsername on cache miss.
func (p *PeerStorage) SetResolver(fn func(ctx context.Context, username string) ([]tg.UserClass, []tg.ChatClass, error)) {
	p.resolveUsernameFn = fn
}

// SetPhoneResolver injects a callback that resolves a phone number via the Telegram API
// (contacts.ResolvePhone RPC). Called automatically by GetPeerByPhoneNumber on cache miss.
func (p *PeerStorage) SetPhoneResolver(fn func(ctx context.Context, phone string) ([]tg.UserClass, []tg.ChatClass, error)) {
	p.resolvePhoneFn = fn
}

// startPruner runs a background goroutine that deletes stale peers from the DB
// every 24 hours. Peers older than 30 days are removed.
func (p *PeerStorage) startPruner() {
	defer close(p.pruneDone)
	ticker := time.NewTicker(pruneInterval)
	defer ticker.Stop()

	for {
		select {
		case <-p.stopPrune:
			return
		case <-ticker.C:
			cutoff := time.Now().Add(-30 * 24 * time.Hour).Unix()
			n, err := p.db.DeleteStalePeers(cutoff)
			if err != nil {
				log.Printf("peers: prune failed: %v", err)
			} else if n > 0 {
				log.Printf("peers: pruned %d stale peers (older than %s)", n, time.Unix(cutoff, 0).Format("2006-01-02"))
			}
		}
	}
}

// Close closes the write channel, waits for the writer and pruner goroutines to drain,
// and closes the adapter to release database connections.
// Safe to call multiple times (idempotent).
func (p *PeerStorage) Close() {
	if p.stopPrune != nil {
		close(p.stopPrune)
		<-p.pruneDone
		p.stopPrune = nil
	}
	if p.writeCh != nil {
		close(p.writeCh)
		<-p.writerDone
		p.writeCh = nil
	}
	if p.db != nil {
		p.db.Close()
		p.db = nil
	}
}
