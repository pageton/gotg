package pgxdb

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pageton/gotg/storage"
)

type PgxAdapter struct {
	pool *pgxpool.Pool
	ctx  context.Context
}

type Option func(*PgxAdapter)

func WithContext(ctx context.Context) Option {
	return func(p *PgxAdapter) { p.ctx = ctx }
}

func New(pool *pgxpool.Pool, opts ...Option) *PgxAdapter {
	a := &PgxAdapter{
		pool: pool,
		ctx:  context.Background(),
	}
	for _, o := range opts {
		o(a)
	}
	return a
}

func NewFromDSN(dsn string, opts ...Option) (*PgxAdapter, error) {
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		return nil, err
	}
	a := New(pool, opts...)
	return a, nil
}

func (p *PgxAdapter) Pool() *pgxpool.Pool { return p.pool }

func (p *PgxAdapter) GetSession(version int) (*storage.Session, error) {
	s := &storage.Session{}
	err := p.pool.QueryRow(p.ctx,
		"SELECT version, data FROM sessions WHERE version = $1", version,
	).Scan(&s.Version, &s.Data)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, nil
		}
		return nil, err
	}
	return s, nil
}

func (p *PgxAdapter) UpdateSession(s *storage.Session) error {
	_, err := p.pool.Exec(p.ctx,
		`INSERT INTO sessions (version, data) VALUES ($1, $2)
		 ON CONFLICT (version) DO UPDATE SET data = EXCLUDED.data`,
		s.Version, s.Data,
	)
	return err
}

func (p *PgxAdapter) SavePeer(peer *storage.Peer) error {
	_, err := p.pool.Exec(p.ctx,
		`INSERT INTO peers (id, access_hash, type, username, language) VALUES ($1, $2, $3, $4, $5)
		 ON CONFLICT (id) DO UPDATE SET access_hash = EXCLUDED.access_hash, type = EXCLUDED.type,
		 username = EXCLUDED.username, language = EXCLUDED.language`,
		peer.ID, peer.AccessHash, peer.Type, peer.Username, peer.Language,
	)
	return err
}

func (p *PgxAdapter) GetPeerByID(id int64) (*storage.Peer, error) {
	peer := &storage.Peer{}
	err := p.pool.QueryRow(p.ctx,
		"SELECT id, access_hash, type, username, language FROM peers WHERE id = $1", id,
	).Scan(&peer.ID, &peer.AccessHash, &peer.Type, &peer.Username, &peer.Language)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, nil
		}
		return nil, err
	}
	return peer, nil
}

func (p *PgxAdapter) GetPeerByUsername(username string) (*storage.Peer, error) {
	peer := &storage.Peer{}
	err := p.pool.QueryRow(p.ctx,
		"SELECT id, access_hash, type, username, language FROM peers WHERE username = $1", username,
	).Scan(&peer.ID, &peer.AccessHash, &peer.Type, &peer.Username, &peer.Language)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, nil
		}
		return nil, err
	}
	return peer, nil
}

func (p *PgxAdapter) SaveConvState(state *storage.ConvState) error {
	_, err := p.pool.Exec(p.ctx,
		`INSERT INTO conv_states (key, chat_id, user_id, step, payload, expires_at, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
		 ON CONFLICT (key) DO UPDATE SET chat_id = EXCLUDED.chat_id, user_id = EXCLUDED.user_id,
		 step = EXCLUDED.step, payload = EXCLUDED.payload, expires_at = EXCLUDED.expires_at, updated_at = NOW()`,
		state.Key, state.ChatID, state.UserID, state.Step, state.Payload, state.ExpiresAt, state.CreatedAt,
	)
	return err
}

func (p *PgxAdapter) LoadConvState(key string) (*storage.ConvState, error) {
	state := &storage.ConvState{}
	err := p.pool.QueryRow(p.ctx,
		"SELECT key, chat_id, user_id, step, payload, expires_at, created_at, updated_at FROM conv_states WHERE key = $1", key,
	).Scan(&state.Key, &state.ChatID, &state.UserID, &state.Step, &state.Payload,
		&state.ExpiresAt, &state.CreatedAt, &state.UpdatedAt)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, nil
		}
		return nil, err
	}
	return state, nil
}

func (p *PgxAdapter) DeleteConvState(key string) error {
	_, err := p.pool.Exec(p.ctx, "DELETE FROM conv_states WHERE key = $1", key)
	return err
}

func (p *PgxAdapter) ListConvStates() ([]storage.ConvState, error) {
	rows, err := p.pool.Query(p.ctx,
		"SELECT key, chat_id, user_id, step, payload, expires_at, created_at, updated_at FROM conv_states",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var states []storage.ConvState
	for rows.Next() {
		var s storage.ConvState
		if err := rows.Scan(&s.Key, &s.ChatID, &s.UserID, &s.Step, &s.Payload,
			&s.ExpiresAt, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		states = append(states, s)
	}
	return states, rows.Err()
}

func (p *PgxAdapter) AutoMigrate() error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS sessions (
			version INTEGER PRIMARY KEY,
			data BYTEA
		)`,
		`CREATE TABLE IF NOT EXISTS peers (
			id BIGINT PRIMARY KEY,
			access_hash BIGINT NOT NULL DEFAULT 0,
			type INTEGER NOT NULL DEFAULT 0,
			username TEXT NOT NULL DEFAULT '',
			language TEXT NOT NULL DEFAULT ''
		)`,
		`CREATE INDEX IF NOT EXISTS idx_peers_username ON peers(username)`,
		`CREATE TABLE IF NOT EXISTS conv_states (
			key TEXT PRIMARY KEY,
			chat_id BIGINT NOT NULL DEFAULT 0,
			user_id BIGINT NOT NULL DEFAULT 0,
			step TEXT NOT NULL DEFAULT '',
			payload BYTEA,
			expires_at TIMESTAMPTZ,
			created_at TIMESTAMPTZ,
			updated_at TIMESTAMPTZ
		)`,
		`CREATE INDEX IF NOT EXISTS idx_conv_states_chat_id ON conv_states(chat_id)`,
		`CREATE INDEX IF NOT EXISTS idx_conv_states_user_id ON conv_states(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_conv_states_expires_at ON conv_states(expires_at)`,
	}
	for _, stmt := range stmts {
		if _, err := p.pool.Exec(p.ctx, stmt); err != nil {
			return err
		}
	}
	return nil
}

func (p *PgxAdapter) Close() error {
	p.pool.Close()
	return nil
}
