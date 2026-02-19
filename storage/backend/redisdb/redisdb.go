// Package redisdb provides a Redis-based Backend implementation.
// Suitable for distributed setups where multiple bot instances share state.
//
// Usage:
//
//	rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
//	b := redisdb.New(rdb)
package redisdb

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/pageton/gotg/storage"
	"github.com/redis/go-redis/v9"
)

const (
	keySession   = "gotg:session:%d"
	keyPeer      = "gotg:peer:%d"
	keyPeerUser  = "gotg:peer:username:%s"
	keyConvState = "gotg:conv:%s"
	keyConvIndex = "gotg:conv:keys"
)

type RedisAdapter struct {
	rdb *redis.Client
	ctx context.Context
}

// Option configures a RedisBackend.
type Option func(*RedisAdapter)

// WithContext sets the context for Redis operations.
func WithContext(ctx context.Context) Option {
	return func(r *RedisAdapter) { r.ctx = ctx }
}

// New creates a new RedisBackend.
func New(rdb *redis.Client, opts ...Option) *RedisAdapter {
	b := &RedisAdapter{
		rdb: rdb,
		ctx: context.Background(),
	}
	for _, o := range opts {
		o(b)
	}
	return b
}

func (r *RedisAdapter) GetSession(version int) (*storage.Session, error) {
	key := fmt.Sprintf(keySession, version)
	data, err := r.rdb.Get(r.ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var s storage.Session
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *RedisAdapter) UpdateSession(s *storage.Session) error {
	key := fmt.Sprintf(keySession, s.Version)
	data, err := json.Marshal(s)
	if err != nil {
		return err
	}
	return r.rdb.Set(r.ctx, key, data, 0).Err()
}

func (r *RedisAdapter) SavePeer(p *storage.Peer) error {
	key := fmt.Sprintf(keyPeer, p.ID)
	data, err := json.Marshal(p)
	if err != nil {
		return err
	}
	pipe := r.rdb.Pipeline()
	pipe.Set(r.ctx, key, data, 0)
	if p.Username != "" {
		pipe.Set(r.ctx, fmt.Sprintf(keyPeerUser, p.Username), fmt.Sprintf("%d", p.ID), 0)
	}
	_, err = pipe.Exec(r.ctx)
	return err
}

func (r *RedisAdapter) GetPeerByID(id int64) (*storage.Peer, error) {
	key := fmt.Sprintf(keyPeer, id)
	data, err := r.rdb.Get(r.ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var p storage.Peer
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *RedisAdapter) GetPeerByUsername(username string) (*storage.Peer, error) {
	idKey := fmt.Sprintf(keyPeerUser, username)
	idStr, err := r.rdb.Get(r.ctx, idKey).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var id int64
	if _, err := fmt.Sscanf(idStr, "%d", &id); err != nil {
		return nil, err
	}
	return r.GetPeerByID(id)
}

func (r *RedisAdapter) SaveConvState(state *storage.ConvState) error {
	key := fmt.Sprintf(keyConvState, state.Key)
	data, err := json.Marshal(state)
	if err != nil {
		return err
	}
	pipe := r.rdb.Pipeline()
	pipe.Set(r.ctx, key, data, 0)
	pipe.SAdd(r.ctx, keyConvIndex, state.Key)
	_, err = pipe.Exec(r.ctx)
	return err
}

func (r *RedisAdapter) LoadConvState(key string) (*storage.ConvState, error) {
	rKey := fmt.Sprintf(keyConvState, key)
	data, err := r.rdb.Get(r.ctx, rKey).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var state storage.ConvState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, err
	}
	return &state, nil
}

func (r *RedisAdapter) DeleteConvState(key string) error {
	rKey := fmt.Sprintf(keyConvState, key)
	pipe := r.rdb.Pipeline()
	pipe.Del(r.ctx, rKey)
	pipe.SRem(r.ctx, keyConvIndex, key)
	_, err := pipe.Exec(r.ctx)
	return err
}

func (r *RedisAdapter) ListConvStates() ([]storage.ConvState, error) {
	keys, err := r.rdb.SMembers(r.ctx, keyConvIndex).Result()
	if err != nil {
		return nil, err
	}
	states := make([]storage.ConvState, 0, len(keys))
	for _, k := range keys {
		s, err := r.LoadConvState(k)
		if err != nil {
			return nil, err
		}
		if s != nil {
			states = append(states, *s)
		}
	}
	return states, nil
}

func (r *RedisAdapter) AutoMigrate() error { return nil }

func (r *RedisAdapter) Close() error {
	return r.rdb.Close()
}
