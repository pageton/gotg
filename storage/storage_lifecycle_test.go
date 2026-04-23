package storage_test

import (
	"testing"
	"time"

	"github.com/pageton/gotg/storage"
	"github.com/pageton/gotg/storage/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPeerStorageDrainReopenCycle(t *testing.T) {
	adapter := memory.New()
	ps, err := storage.NewPeerStorageWithAdapter(adapter, false)
	require.NoError(t, err)

	// Write a peer before draining.
	ps.AddPeer(123, 456, storage.TypeUser, "testuser")

	// Verify peer is accessible.
	peer := ps.GetPeerByID(123)
	require.NotNil(t, peer)
	assert.Equal(t, int64(456), peer.AccessHash)

	// Drain — stops goroutines but keeps adapter alive.
	ps.Drain()
	assert.True(t, ps.IsDrained(), "expected IsDrained after Drain")

	// Reopen — restarts goroutines.
	err = ps.Reopen()
	require.NoError(t, err)
	assert.False(t, ps.IsDrained(), "expected not drained after Reopen")

	// Verify we can still write peers after reopen.
	// Use TypeUser so the TDLibPeerID encoding is identity (no transform),
	// keeping the focus on Drain/Reopen lifecycle, not ID encoding.
	ps.AddPeer(789, 101112, storage.TypeUser, "testuser2")
	peer = ps.GetPeerByID(789)
	require.NotNil(t, peer)
	assert.Equal(t, int64(101112), peer.AccessHash)

	// Old peer should still be accessible (in-memory cache survives Drain).
	peer = ps.GetPeerByID(123)
	require.NotNil(t, peer)
	assert.Equal(t, int64(456), peer.AccessHash)

	// Clean up.
	ps.Close()
}

func TestPeerStorageClosePreventsReopen(t *testing.T) {
	adapter := memory.New()
	ps, err := storage.NewPeerStorageWithAdapter(adapter, false)
	require.NoError(t, err)

	ps.Close()
	assert.True(t, ps.IsDrained())

	err = ps.Reopen()
	assert.Error(t, err, "Reopen after Close should fail because adapter is nil")
	assert.Contains(t, err.Error(), "adapter is nil")
}

func TestPeerStorageDrainIdempotent(t *testing.T) {
	adapter := memory.New()
	ps, err := storage.NewPeerStorageWithAdapter(adapter, false)
	require.NoError(t, err)

	// Double drain should not panic.
	ps.Drain()
	ps.Drain()
	assert.True(t, ps.IsDrained())

	ps.Close()
}

func TestPeerStorageCloseIdempotent(t *testing.T) {
	adapter := memory.New()
	ps, err := storage.NewPeerStorageWithAdapter(adapter, false)
	require.NoError(t, err)

	// Double close should not panic.
	ps.Close()
	ps.Close()
}

func TestPeerStorageInMemorySkipsDrainReopen(t *testing.T) {
	adapter := memory.New()
	ps, err := storage.NewPeerStorageWithAdapter(adapter, true)
	require.NoError(t, err)

	// In-memory storage is never considered drained.
	assert.False(t, ps.IsDrained())

	// Drain is a no-op.
	ps.Drain()
	assert.False(t, ps.IsDrained())

	// Reopen is a no-op and never errors.
	err = ps.Reopen()
	require.NoError(t, err)

	ps.Close()
}

func TestPeerStorageSessionSurvivesDrainReopen(t *testing.T) {
	adapter := memory.New()
	ps, err := storage.NewPeerStorageWithAdapter(adapter, false)
	require.NoError(t, err)

	// Write session data.
	authKey := make([]byte, 256)
	for i := range authKey {
		authKey[i] = byte(i)
	}
	ps.UpdateSession(&storage.Session{Version: storage.LatestVersion, Data: authKey})

	// Verify session is readable.
	s := ps.GetSession()
	require.NotNil(t, s)
	assert.Equal(t, authKey, s.Data)

	// Drain and reopen.
	ps.Drain()
	err = ps.Reopen()
	require.NoError(t, err)

	// Session should still be accessible from the adapter.
	s = ps.GetSession()
	require.NotNil(t, s)
	assert.Equal(t, authKey, s.Data)

	ps.Close()
}

func TestPeerStorageAddPeerAfterDrainReopenDoesNotDeadlock(t *testing.T) {
	adapter := memory.New()
	ps, err := storage.NewPeerStorageWithAdapter(adapter, false)
	require.NoError(t, err)

	ps.AddPeer(100, 200, storage.TypeUser, "before_drain")

	ps.Drain()

	err = ps.Reopen()
	require.NoError(t, err)

	// This would deadlock if writeCh was not recreated.
	done := make(chan struct{})
	go func() {
		defer close(done)
		ps.AddPeer(300, 400, storage.TypeUser, "after_reopen")
	}()

	select {
	case <-done:
		// Success — no deadlock.
	case <-time.After(5 * time.Second):
		t.Fatal("AddPeer deadlocked after Drain/Reopen")
	}

	peer := ps.GetPeerByID(300)
	require.NotNil(t, peer)
	assert.Equal(t, int64(400), peer.AccessHash)

	ps.Close()
}

func TestEncryptedSessionRoundTrip(t *testing.T) {
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}

	enc, err := storage.NewSessionEncryptor(key)
	require.NoError(t, err)

	adapter := memory.New()
	ps, err := storage.NewPeerStorageWithAdapter(adapter, false)
	require.NoError(t, err)
	ps.SetEncryptor(enc)

	// Write session with auth key data.
	authKey := make([]byte, 256)
	for i := range authKey {
		authKey[i] = byte(0xFF - i%256)
	}
	ps.UpdateSession(&storage.Session{Version: storage.LatestVersion, Data: authKey})

	// Read back — should decrypt transparently.
	s := ps.GetSession()
	require.NotNil(t, s)
	assert.Equal(t, authKey, s.Data)

	// Verify raw adapter data is actually encrypted (different from plaintext).
	rawSession, err := adapter.GetSession(storage.LatestVersion)
	require.NoError(t, err)
	require.NotNil(t, rawSession)
	assert.NotEqual(t, authKey, rawSession.Data, "raw adapter data should be encrypted")
	assert.Greater(t, len(rawSession.Data), len(authKey), "encrypted data should be longer (nonce + tag)")

	ps.Close()
}

func TestEncryptedSessionWrongKeyFails(t *testing.T) {
	key1 := make([]byte, 32)
	key2 := make([]byte, 32)
	key2[0] = 0xFF // different key

	enc1, err := storage.NewSessionEncryptor(key1)
	require.NoError(t, err)
	enc2, err := storage.NewSessionEncryptor(key2)
	require.NoError(t, err)

	adapter := memory.New()
	ps, err := storage.NewPeerStorageWithAdapter(adapter, false)
	require.NoError(t, err)

	// Write with key1.
	ps.SetEncryptor(enc1)
	ps.UpdateSession(&storage.Session{Version: storage.LatestVersion, Data: []byte("secret")})

	// Switch to key2 and try to read — should fail gracefully.
	ps.SetEncryptor(enc2)
	s := ps.GetSession()
	// GetSession returns a zero-value session on decrypt failure.
	require.NotNil(t, s)
	assert.Equal(t, storage.LatestVersion, s.Version)
	assert.Empty(t, s.Data, "decrypt with wrong key should return empty data")

	ps.Close()
}

func TestNoEncryptionPassthrough(t *testing.T) {
	adapter := memory.New()
	ps, err := storage.NewPeerStorageWithAdapter(adapter, false)
	require.NoError(t, err)
	// No encryptor set — data should pass through unchanged.

	authKey := []byte("raw session data without encryption")
	ps.UpdateSession(&storage.Session{Version: storage.LatestVersion, Data: authKey})

	s := ps.GetSession()
	require.NotNil(t, s)
	assert.Equal(t, authKey, s.Data)

	// Raw adapter should have the exact same data.
	rawSession, err := adapter.GetSession(storage.LatestVersion)
	require.NoError(t, err)
	assert.Equal(t, authKey, rawSession.Data)

	ps.Close()
}
