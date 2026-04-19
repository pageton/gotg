package mongodb

import (
	"context"
	"errors"

	"github.com/pageton/gotg/session"
	"github.com/pageton/gotg/storage"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const (
	colSessions   = "sessions"
	colPeers      = "peers"
	colConvStates = "conv_states"
)

type MongoAdapter struct {
	db  *mongo.Database
	ctx context.Context
}

type Option func(*MongoAdapter)

func WithContext(ctx context.Context) Option {
	return func(m *MongoAdapter) { m.ctx = ctx }
}

func New(db *mongo.Database, opts ...Option) *MongoAdapter {
	a := &MongoAdapter{
		db:  db,
		ctx: context.Background(),
	}
	for _, o := range opts {
		o(a)
	}
	return a
}

func NewFromURI(uri, dbName string, opts ...Option) (*MongoAdapter, error) {
	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}
	return New(client.Database(dbName), opts...), nil
}

func Session(db *mongo.Database, opts ...Option) *session.AdapterSessionConstructor {
	return session.Adapter(New(db, opts...))
}

func SessionFromURI(uri, dbName string, opts ...Option) *session.AdapterSessionConstructor {
	adapter, err := NewFromURI(uri, dbName, opts...)
	if err != nil {
		panic("mongodb: failed to open MongoDB: " + err.Error())
	}
	return session.Adapter(adapter)
}

func (m *MongoAdapter) Database() *mongo.Database { return m.db }

func (m *MongoAdapter) GetSession(version int) (*storage.Session, error) {
	var s storage.Session
	err := m.db.Collection(colSessions).FindOne(m.ctx, bson.M{"version": version}).Decode(&s)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (m *MongoAdapter) UpdateSession(s *storage.Session) error {
	_, err := m.db.Collection(colSessions).ReplaceOne(
		m.ctx,
		bson.M{"version": s.Version},
		s,
		options.Replace().SetUpsert(true),
	)
	return err
}

func (m *MongoAdapter) SavePeer(p *storage.Peer) error {
	_, err := m.db.Collection(colPeers).ReplaceOne(
		m.ctx,
		bson.M{"id": p.ID},
		p,
		options.Replace().SetUpsert(true),
	)
	return err
}

func (m *MongoAdapter) GetPeerByID(id int64) (*storage.Peer, error) {
	var p storage.Peer
	err := m.db.Collection(colPeers).FindOne(m.ctx, bson.M{"id": id}).Decode(&p)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (m *MongoAdapter) GetPeerByUsername(username string) (*storage.Peer, error) {
	var p storage.Peer
	err := m.db.Collection(colPeers).FindOne(m.ctx, bson.M{"username": username}).Decode(&p)
	if err == nil {
		return &p, nil
	}
	if !errors.Is(err, mongo.ErrNoDocuments) {
		return nil, err
	}
	// Search in usernames array.
	err = m.db.Collection(colPeers).FindOne(m.ctx, bson.M{"usernames": username}).Decode(&p)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (m *MongoAdapter) GetPeerByPhoneNumber(phone string) (*storage.Peer, error) {
	var p storage.Peer
	err := m.db.Collection(colPeers).FindOne(m.ctx, bson.M{"phonenumber": phone}).Decode(&p)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (m *MongoAdapter) SaveConvState(state *storage.ConvState) error {
	_, err := m.db.Collection(colConvStates).ReplaceOne(
		m.ctx,
		bson.M{"key": state.Key},
		state,
		options.Replace().SetUpsert(true),
	)
	return err
}

func (m *MongoAdapter) LoadConvState(key string) (*storage.ConvState, error) {
	var state storage.ConvState
	err := m.db.Collection(colConvStates).FindOne(m.ctx, bson.M{"key": key}).Decode(&state)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &state, nil
}

func (m *MongoAdapter) DeleteConvState(key string) error {
	_, err := m.db.Collection(colConvStates).DeleteOne(m.ctx, bson.M{"key": key})
	return err
}

func (m *MongoAdapter) ListConvStates() ([]storage.ConvState, error) {
	cursor, err := m.db.Collection(colConvStates).Find(m.ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	var states []storage.ConvState
	if err := cursor.All(m.ctx, &states); err != nil {
		return nil, err
	}
	return states, nil
}

func (m *MongoAdapter) AutoMigrate() error {
	col := m.db.Collection(colPeers)
	_, err := col.Indexes().CreateMany(m.ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "username", Value: 1}}, Options: options.Index().SetSparse(true)},
		{Keys: bson.D{{Key: "phonenumber", Value: 1}}, Options: options.Index().SetSparse(true)},
	})
	if err != nil {
		return err
	}

	col = m.db.Collection(colConvStates)
	_, err = col.Indexes().CreateMany(m.ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "chat_id", Value: 1}}},
		{Keys: bson.D{{Key: "user_id", Value: 1}}},
		{Keys: bson.D{{Key: "expires_at", Value: 1}}},
	})
	return err
}

func (m *MongoAdapter) DeleteStalePeers(olderThan int64) (int64, error) {
	filter := bson.M{
		"lastupdated": bson.M{"$gt": 0, "$lt": olderThan},
	}
	result, err := m.db.Collection(colPeers).DeleteMany(m.ctx, filter)
	if err != nil {
		return 0, err
	}
	return result.DeletedCount, nil
}

func (m *MongoAdapter) Close() error {
	return m.db.Client().Disconnect(m.ctx)
}
