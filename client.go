package gotg

//go:generate go run ./generator

import (
	"context"
	"time"

	tdSession "github.com/gotd/td/session"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/telegram/dcs"
	"github.com/gotd/td/tg"
	"github.com/pageton/gotg/conv"
	"github.com/pageton/gotg/dispatcher"
	"github.com/pageton/gotg/session"
	"github.com/pageton/gotg/storage"
	"go.uber.org/zap"
)

const VERSION = "v1.0.0-beta23"

type Client struct {
	// Dispatcher handlers the incoming updates and execute mapped handlers. It is recommended to use dispatcher.MakeDispatcher function for this field.
	Dispatcher  dispatcher.Dispatcher
	ConvManager *conv.Manager
	// PublicKeys of telegram.
	//
	// If not provided, embedded public keys will be used.
	PublicKeys []telegram.PublicKey
	// DC ID to connect.
	//
	// If not provided, 2 will be used by default.
	DC int
	// DCList is initial list of addresses to connect.
	DCList dcs.List
	// Resolver to use.
	Resolver dcs.Resolver
	// MigrationTimeout configures migration timeout.
	MigrationTimeout time.Duration
	// AckBatchSize is limit of MTProto ACK buffer size.
	AckBatchSize int
	// AckInterval is maximum time to buffer MTProto ACK.
	AckInterval time.Duration
	// RetryInterval is duration between send retries.
	RetryInterval time.Duration
	// MaxRetries is limit of send retries.
	MaxRetries int
	// ExchangeTimeout is timeout of every key exchange request.
	ExchangeTimeout time.Duration
	// DialTimeout is timeout of creating connection.
	DialTimeout time.Duration
	// CompressThreshold is a threshold in bytes to determine that message
	// is large enough to be compressed using GZIP.
	// If < 0, compression will be disabled.
	// If == 0, default value will be used.
	CompressThreshold int
	// Whether to show the copyright line in console or no.
	DisableCopyright bool
	// Logger is instance of zap.Logger. No logs by default.
	Logger *zap.Logger
	// Session info of the authenticated user, use session.NewSession function to fill this field.
	sessionStorage tdSession.Storage
	// Self contains details of logged in user in the form of *tg.User.
	Self *tg.User
	// Code for the language used on the device's OS, ISO 639-1 standard.
	SystemLangCode string
	// Code for the language used on the client, ISO 639-1 standard.
	ClientLangCode string
	// PeerStorage is the storage for all the peers.
	// It is recommended to use storage.NewPeerStorage function for this field.
	PeerStorage *storage.PeerStorage
	// NoAutoAuth is a flag to disable automatic authentication
	// if the current session is invalid.
	NoAutoAuth bool
	// NoUpdates is a flag to disable updates.
	NoUpdates bool

	authConversator AuthConversator
	sendCodeOptions auth.SendCodeOptions
	clientType      clientType
	ctx             context.Context
	err             error
	autoFetchReply  bool
	cancel          context.CancelFunc
	running         bool
	*telegram.Client
	apiID   int
	apiHash string
}

type ClientOpts struct {
	// Logger is instance of zap.Logger. No logs by default.
	Logger *zap.Logger
	// Whether to store session and peer storage in memory or not
	//
	// Note: Sessions and Peers won't be persistent if this field is set to true.
	InMemory bool
	// PublicKeys of telegram.
	//
	// If not provided, embedded public keys will be used.
	PublicKeys []telegram.PublicKey
	// DC ID to connect.
	//
	// If not provided, 2 will be used by default.
	DC int
	// DCList is initial list of addresses to connect.
	DCList dcs.List
	// Resolver to use.
	Resolver dcs.Resolver
	// Whether to show the copyright line in console or no.
	DisableCopyright bool
	// Session info of the authenticated user, use session.NewSession function to fill this field.
	Session session.SessionConstructor
	// Setting this field to true will lead to automatically fetch the reply_to_message for a new message update.
	//
	// Set to `false` by default.
	AutoFetchReply bool
	// Setting this field to true will lead to automatically fetch the entire reply_to_message chain for a new message update.
	//
	// Set to `false` by default.
	FetchEntireReplyChain bool
	// Code for the language used on the device's OS, ISO 639-1 standard.
	SystemLangCode string
	// Code for the language used on the client, ISO 639-1 standard.
	ClientLangCode string
	// Custom client device
	Device *telegram.DeviceConfig
	// Panic handles all the panics that occur during handler execution.
	PanicHandler dispatcher.PanicHandler
	// Error handles all the unknown errors which are returned by the handler callback functions.
	ErrorHandler dispatcher.ErrorHandler
	// Custom Middlewares
	Middlewares []telegram.Middleware
	// DispatcherMiddlewares are middleware handlers that run before user-added handlers.
	// These are automatically added to groups 0, 1, 2, ... in order.
	// Useful for i18n, authentication, logging, etc.
	DispatcherMiddlewares []dispatcher.Handler
	// Custom Run() Middleware
	// Can be used for floodWaiter package
	// https://github.com/pageton/gotg/blob/beta/examples/middleware/main.go#L41
	RunMiddleware func(
		origRun func(ctx context.Context, f func(ctx context.Context) error) (err error),
		ctx context.Context,
		f func(ctx context.Context) (err error),
	) (err error)
	// A custom context to use for the client.
	// If not provided, context.Background() will be used.
	// Note: This context will be used for the entire lifecycle of the client.
	Context context.Context
	// AuthConversator is the interface for the authenticator.
	// gotg.BasicConversator is used by default.
	AuthConversator AuthConversator
	// MigrationTimeout configures migration timeout.
	MigrationTimeout time.Duration
	// AckBatchSize is limit of MTProto ACK buffer size.
	AckBatchSize int
	// AckInterval is maximum time to buffer MTProto ACK.
	AckInterval time.Duration
	// RetryInterval is duration between send retries.
	RetryInterval time.Duration
	// MaxRetries is limit of send retries.
	MaxRetries int
	// ExchangeTimeout is timeout of every key exchange request.
	ExchangeTimeout time.Duration
	// DialTimeout is timeout of creating connection.
	DialTimeout time.Duration
	// CompressThreshold is a threshold in bytes to determine that message
	// is large enough to be compressed using GZIP.
	// If < 0, compression will be disabled.
	// If == 0, default value will be used.
	CompressThreshold int
	// NoAutoAuth is a flag to disable automatic authentication
	// if the current session is invalid.
	NoAutoAuth bool
	// NoUpdates is a flag to disable updates.
	NoUpdates bool
	// SendCodeOptions allows overriding AuthSendCode behavior.
	SendCodeOptions *auth.SendCodeOptions
	// Only usable by Users not bots
	// PeersFromDialogs is a flag to enable adding peers fetched
	// from dialogs to memory/database on startup
	PeersFromDialogs bool
	// WaitOnPeersFromDialogs is a flag to enable waiting on
	// PeersFromDialogs to complete during client start
	WaitOnPeersFromDialogs bool
}

// NewClient creates a new gotg client and logs in to telegram.
func NewClient(apiID int, apiHash string, clientType clientType, opts *ClientOpts) (*Client, error) {
	if opts == nil {
		opts = &ClientOpts{
			SystemLangCode: "en",
			ClientLangCode: "en",
		}
	}

	if opts.Context == nil {
		opts.Context = context.Background()
	}
	ctx, cancel := context.WithCancel(opts.Context)

	peerStorage, sessionStorage, err := session.NewSessionStorage(ctx, opts.Session, opts.InMemory)
	if err != nil {
		cancel()
		return nil, err
	}

	if opts.AuthConversator == nil {
		opts.AuthConversator = BasicConversator()
	}

	d := dispatcher.NewNativeDispatcher(opts.AutoFetchReply, opts.FetchEntireReplyChain, opts.ErrorHandler, opts.PanicHandler, peerStorage)

	for i, middleware := range opts.DispatcherMiddlewares {
		d.AddHandlerToGroup(middleware, i)
	}

	c := Client{
		Resolver:          opts.Resolver,
		PublicKeys:        opts.PublicKeys,
		DC:                opts.DC,
		DCList:            opts.DCList,
		MigrationTimeout:  opts.MigrationTimeout,
		AckBatchSize:      opts.AckBatchSize,
		AckInterval:       opts.AckInterval,
		RetryInterval:     opts.RetryInterval,
		MaxRetries:        opts.MaxRetries,
		ExchangeTimeout:   opts.ExchangeTimeout,
		DialTimeout:       opts.DialTimeout,
		CompressThreshold: opts.CompressThreshold,
		DisableCopyright:  opts.DisableCopyright,
		Logger:            opts.Logger,
		SystemLangCode:    opts.SystemLangCode,
		ClientLangCode:    opts.ClientLangCode,
		NoAutoAuth:        opts.NoAutoAuth,
		NoUpdates:         opts.NoUpdates,
		authConversator:   opts.AuthConversator,
		Dispatcher:        d,
		ConvManager:       d.ConvManager(),
		PeerStorage:       peerStorage,
		sessionStorage:    sessionStorage,
		clientType:        clientType,
		ctx:               ctx,
		autoFetchReply:    opts.AutoFetchReply,
		cancel:            cancel,
		apiID:             apiID,
		apiHash:           apiHash,
	}

	if opts.SendCodeOptions != nil {
		c.sendCodeOptions = *opts.SendCodeOptions
	}

	c.printCredit()

	return &c, c.Start(opts)
}
