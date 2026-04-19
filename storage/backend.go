package storage

type Adapter interface {
	GetSession(version int) (*Session, error)
	UpdateSession(s *Session) error

	SavePeer(p *Peer) error
	GetPeerByID(id int64) (*Peer, error)
	GetPeerByUsername(username string) (*Peer, error)
	GetPeerByPhoneNumber(phone string) (*Peer, error)

	SaveConvState(state *ConvState) error
	LoadConvState(key string) (*ConvState, error)
	DeleteConvState(key string) error
	ListConvStates() ([]ConvState, error)

	AutoMigrate() error
	Close() error

	// DeleteStalePeers removes peers whose last_updated is older than the
	// given unix timestamp. Returns the number of deleted peers.
	DeleteStalePeers(olderThan int64) (int64, error)
}
