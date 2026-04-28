// Package gotg provides a high-level Telegram MTProto framework for building bots and userbots.
// Built on top of gotd/td, it offers intuitive APIs for session management, update handling,
// media operations, conversations, i18n, and Telegram Business API support.

package gotg

//go:generate go run ./generator

import (
	"context"
	"fmt"
	"time"

	tdSession "github.com/gotd/td/session"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/telegram/dcs"
	"github.com/gotd/td/telegram/updates"
	"github.com/gotd/td/tg"
	"github.com/pageton/gotg/conv"
	"github.com/pageton/gotg/dispatcher"
	gotglog "github.com/pageton/gotg/log"
	"github.com/pageton/gotg/session"
	"github.com/pageton/gotg/storage"
)

const VERSION = "v1.0.0-beta25"

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
	Logger           *gotglog.Logger
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
	outgoing        bool
	cancel          context.CancelFunc
	running         bool
	*telegram.Client
	gapManager       *updates.Manager
	apiID            int
	apiHash          string
	deviceParams     tg.JSONValueClass
	defaultParseMode  string
	autoReconnect     *ReconnectConfig
	dialogsDone       chan struct{} // closed when AddPeersFromDialogs goroutine exits
	startOpts         *ClientOpts  // preserved from first Start() for reconnection
}

type ClientOpts struct {
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
	// SendOutgoing enables synthetic outgoing updates for send/edit/delete.
	// When true, sent messages are re-dispatched through handlers with Out=true.
	SendOutgoing bool
	// DisableAutoStart prevents NewClient from starting the client automatically.
	// When true, call Client.Start(opts) manually.
	DisableAutoStart bool
	// LogConfig configures the built-in gotg logger attached to each Update.
	// If nil, DefaultConfig() is used (info level, color, timestamps, no caller).
	LogConfig *gotglog.Config
	// SendCodeOptions allows overriding AuthSendCode behavior.
	SendCodeOptions *auth.SendCodeOptions
	// Only usable by Users not bots
	// PeersFromDialogs is a flag to enable adding peers fetched
	// from dialogs to memory/database on startup
	PeersFromDialogs bool
	// WaitOnPeersFromDialogs is a flag to enable waiting on
	// PeersFromDialogs to complete during client start
	WaitOnPeersFromDialogs bool
	// MaxConcurrentUpdates limits parallel update handler goroutines.
	// Default 1000 when 0.
	MaxConcurrentUpdates int
	// ParseMode sets the default parse mode for all message sending methods.
	// If empty, no formatting is applied unless explicitly specified per-method.
	// Supported values: "HTML", "Markdown", "MarkdownV2", or "" (none).
	ParseMode string
	// AutoReconnect enables automatic reconnection on disconnection.
	// When set, call Client.RunForever() instead of Client.Idle() to use
	// the reconnection loop. See ReconnectConfig for configuration options.
	AutoReconnect *ReconnectConfig
	// SessionEncryptionKey enables AES-256-GCM encryption of session data
	// (the 256-byte MTProto auth key) before persisting to storage. Must be
	// exactly 32 bytes. Source from a secure secret store (env var,
	// /run/secrets, KMS, etc.).
	//
	// WARNING: Once enabled, the database cannot be read without the same
	// key. Back up your plaintext session before enabling.
	SessionEncryptionKey []byte
}

// NewClient creates a new gotg client and authenticates with Telegram.
// It initializes the MTProto client, sets up session storage, peers, dispatcher,
// and conversation manager. For bots, use AsBot(token); for users, use AsUser(phone).
//
// Parameters:
//   - apiID: Telegram API ID from https://my.telegram.org/apps
//   - apiHash: Telegram API Hash from https://my.telegram.org/apps
//   - clientType: Authentication type (AsBot/AsUser/Simple)
//   - opts: Optional client configuration (sessions, logging, middlewares)
//
// Returns:
//   - *Client: Authenticated client ready to start receiving updates
//   - error: Nil on success, or error during session initialization/auth
//
// Errors:
//   - Session initialization failure (invalid storage, encryption error)
//   - Authentication failure (invalid API credentials, network issues)
//   - Dispatcher setup failure
//
// Example:
//
//	client, err := gotg.NewClient(
//	    123456, "your_api_hash",
//	    gotg.AsBot("123456:ABC-DEF..."),
//	    &gotg.ClientOpts{InMemory: true},
//	)
//	if err != nil {
//	    log.Fatalf("client creation failed: %v", err)
//	}
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

	// Enable session encryption if a key is provided.
	if len(opts.SessionEncryptionKey) > 0 {
		enc, encErr := storage.NewSessionEncryptor(opts.SessionEncryptionKey)
		if encErr != nil {
			cancel()
			return nil, fmt.Errorf("session encryption: %w", encErr)
		}
		peerStorage.SetEncryptor(enc)
	}

	if opts.AuthConversator == nil {
		opts.AuthConversator = BasicConversator()
	}

	var logger *gotglog.Logger
	if opts.LogConfig != nil {
		logger = gotglog.New(*opts.LogConfig)
	} else {
		logger = gotglog.Nop()
	}

	d := dispatcher.NewNativeDispatcher(opts.AutoFetchReply, opts.FetchEntireReplyChain, opts.ErrorHandler, opts.PanicHandler, peerStorage, logger, opts.SendOutgoing, opts.ParseMode)
	if opts.MaxConcurrentUpdates > 0 {
		d.SetMaxConcurrentUpdates(opts.MaxConcurrentUpdates)
	}

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
		Logger:            logger,
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
		outgoing:          opts.SendOutgoing,
		cancel:            cancel,
		apiID:             apiID,
		apiHash:           apiHash,
		defaultParseMode:  opts.ParseMode,
		autoReconnect:     opts.AutoReconnect,
	}

	if opts.SendCodeOptions != nil {
		c.sendCodeOptions = *opts.SendCodeOptions
	}

	c.printCredit()

	// Preserve opts for RunForever reconnection (Device, Middlewares, RunMiddleware, etc).
	c.startOpts = opts

	if opts.DisableAutoStart {
		return &c, nil
	}

	return &c, c.Start(opts)
}
