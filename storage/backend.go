package storage

type Adapter interface {
	GetSession(version int) (*Session, error)
	UpdateSession(s *Session) error

	SavePeer(p *Peer) error
	GetPeerByID(id int64) (*Peer, error)
	GetPeerByUsername(username string) (*Peer, error)

	SaveConvState(state *ConvState) error
	LoadConvState(key string) (*ConvState, error)
	DeleteConvState(key string) error
	ListConvStates() ([]ConvState, error)

	AutoMigrate() error
	Close() error
}
