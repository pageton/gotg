package sqlitedb

import (
	"context"
	"database/sql"
	"errors"
	"sync"

	"github.com/bytedance/sonic"
	"github.com/pageton/gotg/session"
	"github.com/pageton/gotg/storage"
	_ "modernc.org/sqlite"
)

type SQLiteAdapter struct {
	db     *sql.DB
	ctx    context.Context
	name   string
	closed bool
	mu     sync.Mutex
}

type Option func(*SQLiteAdapter)

func Context(ctx context.Context) Option {
	return func(s *SQLiteAdapter) { s.ctx = ctx }
}

// WithContext sets the adapter context.
// Deprecated: use Context instead.
func WithContext(ctx context.Context) Option { return Context(ctx) }

func SessionName(name string) Option {
	return func(s *SQLiteAdapter) {
		if name != "" {
			s.name = name
		}
	}
}

func New(db *sql.DB, opts ...Option) *SQLiteAdapter {
	a := &SQLiteAdapter{
		db:   db,
		ctx:  context.Background(),
		name: "default",
	}
	for _, o := range opts {
		o(a)
	}
	return a
}

func NewFromDSN(dsn string, opts ...Option) (*SQLiteAdapter, error) {
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return New(db, opts...), nil
}

func (s *SQLiteAdapter) DB() *sql.DB { return s.db }

func (s *SQLiteAdapter) GetSession(version int) (*storage.Session, error) {
	out := &storage.Session{}
	err := s.db.QueryRowContext(s.ctx,
		"SELECT version, data FROM sessions WHERE session_name = ? AND version = ?", s.name, version,
	).Scan(&out.Version, &out.Data)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return out, nil
}

func (s *SQLiteAdapter) UpdateSession(in *storage.Session) error {
	_, err := s.db.ExecContext(s.ctx,
		`INSERT INTO sessions (session_name, version, data) VALUES (?, ?, ?)
		 ON CONFLICT(session_name, version) DO UPDATE SET data = excluded.data`,
		s.name, in.Version, in.Data,
	)
	return err
}

func (s *SQLiteAdapter) SavePeer(peer *storage.Peer) error {
	usernamesJSON, err := sonic.Marshal(peer.Usernames)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(s.ctx,
		`INSERT INTO peers (session_name, id, access_hash, type, username, usernames, phone_number, is_bot, photo_id, language, last_updated)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		 ON CONFLICT(session_name, id) DO UPDATE SET
		 access_hash = excluded.access_hash,
		 type = excluded.type,
		 username = excluded.username,
		 usernames = excluded.usernames,
		 phone_number = excluded.phone_number,
		 is_bot = excluded.is_bot,
		 photo_id = excluded.photo_id,
		 language = excluded.language,
		 last_updated = excluded.last_updated`,
		s.name, peer.ID, peer.AccessHash, peer.Type, peer.Username, string(usernamesJSON), peer.PhoneNumber, peer.IsBot, peer.PhotoID, peer.Language, peer.LastUpdated,
	)
	return err
}

const peerCols = "id, access_hash, type, username, COALESCE(usernames, '[]'), COALESCE(phone_number, ''), COALESCE(is_bot, 0), COALESCE(photo_id, 0), COALESCE(language, ''), COALESCE(last_updated, 0)"

func scanPeer(row scanner) (*storage.Peer, error) {
	peer := &storage.Peer{}
	var usernamesJSON string
	err := row.Scan(&peer.ID, &peer.AccessHash, &peer.Type, &peer.Username, &usernamesJSON, &peer.PhoneNumber, &peer.IsBot, &peer.PhotoID, &peer.Language, &peer.LastUpdated)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	if usernamesJSON != "" {
		_ = sonic.Unmarshal([]byte(usernamesJSON), &peer.Usernames)
	}
	return peer, nil
}

type scanner interface {
	Scan(dest ...any) error
}

func (s *SQLiteAdapter) GetPeerByID(id int64) (*storage.Peer, error) {
	return scanPeer(s.db.QueryRowContext(s.ctx,
		"SELECT "+peerCols+" FROM peers WHERE session_name = ? AND id = ?", s.name, id,
	))
}

func (s *SQLiteAdapter) GetPeerByUsername(username string) (*storage.Peer, error) {
	peer, err := scanPeer(s.db.QueryRowContext(s.ctx,
		"SELECT "+peerCols+" FROM peers WHERE session_name = ? AND username = ?", s.name, username,
	))
	if err != nil || peer != nil {
		return peer, err
	}
	return scanPeer(s.db.QueryRowContext(s.ctx,
		"SELECT "+peerCols+" FROM peers WHERE session_name = ? AND usernames LIKE ?", s.name, `%"Username":"`+username+`"%`,
	))
}

func (s *SQLiteAdapter) GetPeerByPhoneNumber(phone string) (*storage.Peer, error) {
	return scanPeer(s.db.QueryRowContext(s.ctx,
		"SELECT "+peerCols+" FROM peers WHERE session_name = ? AND phone_number = ?", s.name, phone,
	))
}

func (s *SQLiteAdapter) SaveConvState(state *storage.ConvState) error {
	_, err := s.db.ExecContext(s.ctx,
		`INSERT INTO conv_states (session_name, key, chat_id, user_id, step, payload, expires_at, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		 ON CONFLICT(session_name, key) DO UPDATE SET
		 chat_id = excluded.chat_id,
		 user_id = excluded.user_id,
		 step = excluded.step,
		 payload = excluded.payload,
		 expires_at = excluded.expires_at,
		 created_at = excluded.created_at,
		 updated_at = excluded.updated_at`,
		s.name, state.Key, state.ChatID, state.UserID, state.Step, state.Payload, state.ExpiresAt, state.CreatedAt, state.UpdatedAt,
	)
	return err
}

func (s *SQLiteAdapter) LoadConvState(key string) (*storage.ConvState, error) {
	out := &storage.ConvState{}
	err := s.db.QueryRowContext(s.ctx,
		"SELECT key, chat_id, user_id, step, payload, expires_at, created_at, updated_at FROM conv_states WHERE session_name = ? AND key = ?", s.name, key,
	).Scan(&out.Key, &out.ChatID, &out.UserID, &out.Step, &out.Payload, &out.ExpiresAt, &out.CreatedAt, &out.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return out, nil
}

func (s *SQLiteAdapter) DeleteConvState(key string) error {
	_, err := s.db.ExecContext(s.ctx, "DELETE FROM conv_states WHERE session_name = ? AND key = ?", s.name, key)
	return err
}

func (s *SQLiteAdapter) ListConvStates() ([]storage.ConvState, error) {
	rows, err := s.db.QueryContext(s.ctx,
		"SELECT key, chat_id, user_id, step, payload, expires_at, created_at, updated_at FROM conv_states WHERE session_name = ?", s.name,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var states []storage.ConvState
	for rows.Next() {
		var st storage.ConvState
		if err := rows.Scan(&st.Key, &st.ChatID, &st.UserID, &st.Step, &st.Payload, &st.ExpiresAt, &st.CreatedAt, &st.UpdatedAt); err != nil {
			return nil, err
		}
		states = append(states, st)
	}
	return states, rows.Err()
}

func (s *SQLiteAdapter) AutoMigrate() error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS sessions (
			session_name TEXT NOT NULL DEFAULT 'default',
			version INTEGER NOT NULL,
			data BLOB,
			PRIMARY KEY (session_name, version)
		)`,
		`CREATE TABLE IF NOT EXISTS peers (
			session_name TEXT NOT NULL DEFAULT 'default',
			id INTEGER NOT NULL,
			access_hash INTEGER NOT NULL DEFAULT 0,
			type INTEGER NOT NULL DEFAULT 0,
			username TEXT NOT NULL DEFAULT '',
			usernames TEXT NOT NULL DEFAULT '[]',
			phone_number TEXT NOT NULL DEFAULT '',
			is_bot INTEGER NOT NULL DEFAULT 0,
			photo_id INTEGER NOT NULL DEFAULT 0,
			language TEXT NOT NULL DEFAULT '',
			last_updated INTEGER NOT NULL DEFAULT 0,
			PRIMARY KEY (session_name, id)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_peers_username ON peers(session_name, username)`,
		`CREATE INDEX IF NOT EXISTS idx_peers_phone_number ON peers(session_name, phone_number)`,
		`CREATE TABLE IF NOT EXISTS conv_states (
			session_name TEXT NOT NULL DEFAULT 'default',
			key TEXT NOT NULL,
			chat_id INTEGER NOT NULL DEFAULT 0,
			user_id INTEGER NOT NULL DEFAULT 0,
			step TEXT NOT NULL DEFAULT '',
			payload BLOB,
			expires_at TIMESTAMP,
			created_at TIMESTAMP,
			updated_at TIMESTAMP,
			PRIMARY KEY (session_name, key)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_conv_states_chat_id ON conv_states(session_name, chat_id)`,
		`CREATE INDEX IF NOT EXISTS idx_conv_states_user_id ON conv_states(session_name, user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_conv_states_expires_at ON conv_states(session_name, expires_at)`,
	}
	for _, stmt := range stmts {
		if _, err := s.db.ExecContext(s.ctx, stmt); err != nil {
			return err
		}
	}
	return nil
}

func (s *SQLiteAdapter) DeleteStalePeers(olderThan int64) (int64, error) {
	res, err := s.db.ExecContext(s.ctx,
		"DELETE FROM peers WHERE session_name = ? AND last_updated > 0 AND last_updated < ?", s.name, olderThan,
	)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func (s *SQLiteAdapter) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return nil
	}
	s.closed = true
	// Checkpoint WAL so .wal/.shm files are removed on close.
	s.db.ExecContext(s.ctx, "PRAGMA wal_checkpoint(TRUNCATE)")
	return s.db.Close()
}

func Session(dsn string, opts ...Option) *session.AdapterSessionConstructor {
	adapter, err := NewFromDSN(dsn, opts...)
	if err != nil {
		panic("sqlitedb: failed to open SQLite: " + err.Error())
	}
	return session.Adapter(adapter)
}

// SqliteSession creates a session constructor backed by SQLite.
// Deprecated: use Session instead.
func SqliteSession(dsn string, opts ...Option) *session.AdapterSessionConstructor {
	return Session(dsn, opts...)
}
