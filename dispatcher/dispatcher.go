package dispatcher

import (
	"context"
	"errors"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/tg"
	"github.com/pageton/gotg/adapter"
	"github.com/pageton/gotg/conv"
	gotglog "github.com/pageton/gotg/log"
	"github.com/pageton/gotg/storage"
)

var (
	// StopClient cancels the context and stops the client if returned through handler callback function.
	StopClient = errors.New("disconnect")

	// EndGroups stops iterating over handlers groups if returned through handler callback function.
	EndGroups = errors.New("stopped")
	// ContinueGroups continues iterating over handlers groups if returned through handler callback function.
	ContinueGroups = errors.New("continued")
	// SkipCurrentGroup skips current group and continues iterating over handlers groups if returned through handler callback function.
	SkipCurrentGroup = errors.New("skipped")
)

type Dispatcher interface {
	Initialize(context.Context, context.CancelFunc, *telegram.Client, *tg.User)
	Handle(context.Context, tg.UpdatesClass) error
	AddHandler(Handler)
	AddHandlerToGroup(Handler, int)
	AddHandlers(...Handler)
	AddHandlersToGroup(int, ...Handler)
}

type NativeDispatcher struct {
	cancel              context.CancelFunc
	client              *tg.Client
	self                *tg.User
	sender              *message.Sender
	setReply            bool
	setEntireReplyChain bool
	outgoing            bool
	conv                *conv.Manager
	logger              *gotglog.Logger
	Panic               PanicHandler
	Error               ErrorHandler
	handlerMap          map[int][]Handler
	handlerGroups       []int
	handlerGroupsMu     sync.RWMutex
	pStorage            *storage.PeerStorage
	initwg              *sync.WaitGroup
	nextGroup           atomic.Int64
	updateWg            sync.WaitGroup
	pendingOutgoing     sync.Map
	updateSem           chan struct{}
}

type (
	PanicHandler func(*adapter.Context, *adapter.Update, string)
	ErrorHandler func(*adapter.Context, *adapter.Update, string) error
)

// NewNativeDispatcher creates a new native dispatcher for handling Telegram updates.
//
// The dispatcher manages update processing, handler registration and dispatching.
//
// Parameters:
//   - setReply: Whether to set reply information for messages
//   - setEntireReplyChain: Whether to set full reply chain (forwarded messages)
//   - eHandler: Handler for unknown errors
//   - pHandler: Handler for panic recovery
//   - p: Peer storage for caching peers
//
// Returns:
//   - A new NativeDispatcher instance configured and ready to use
//
// Example:
//
//	dp := dispatcher.NewNativeDispatcher(
//	    true,   // set reply info
//	    true,   // set reply chain
//	    nil,   // default error handler
//	    nil,   // default panic handler
//	    peerStorage,
//	)
func NewNativeDispatcher(setReply bool, setEntireReplyChain bool, eHandler ErrorHandler, pHandler PanicHandler, p *storage.PeerStorage, logger *gotglog.Logger, outgoing bool) *NativeDispatcher {
	if eHandler == nil {
		eHandler = defaultErrorHandler
	}
	if logger == nil {
		logger = gotglog.Default()
	}
	nd := &NativeDispatcher{
		pStorage:            p,
		handlerMap:          make(map[int][]Handler),
		handlerGroups:       make([]int, 0, 8),
		setReply:            setReply,
		setEntireReplyChain: setEntireReplyChain,
		outgoing:            outgoing,
		Error:               eHandler,
		Panic:               pHandler,
		logger:              logger,
		initwg:              &sync.WaitGroup{},
		updateSem:           make(chan struct{}, 1000),
	}
	nd.conv = conv.NewManager(p, 1*time.Minute)
	nd.initwg.Add(1)
	return nd
}

// NewNativeDispatcherWithInit returns a partially initialized dispatcher.
// The WaitGroup will be in a "wait state" after this function.
// Handlers can be added after this function returns, but dispatcher must be
// fully initialized via Initialize() before processing updates (Issue #120).
//
// This pattern is used by AddHandler to wait for initialization completion
// before incrementing the group counter, preventing the race condition.
//
// Returns:
//   - A NativeDispatcher with initwg.Wait() set to 1
//   - A NativeDispatcher struct (not yet fully initialized)
//
// Example:
//
//	// BAD: This causes race condition (Issue #120)
//	dp := NewNativeDispatcher(...)
//	dp.AddHandler(handler)  // Counter increments while initwg.Wait() still running
//
//	// GOOD: This prevents the race (Issue #120)
//	dp, initWg := NewNativeDispatcherWithInit(...)
//	dispatcher.Initialize(dp.initwg, ...)  // Waits for initwg
//	dp.AddHandler(handler)  // Safe - initwg.Wait() is done
func NewNativeDispatcherWithInit(setReply bool, setEntireReplyChain bool, eHandler ErrorHandler, pHandler PanicHandler, p *storage.PeerStorage, logger *gotglog.Logger, outgoing bool, initwg *sync.WaitGroup) *NativeDispatcher {
	if eHandler == nil {
		eHandler = defaultErrorHandler
	}
	nd := NewNativeDispatcher(setReply, setEntireReplyChain, eHandler, pHandler, p, logger, outgoing)
	nd.initwg = initwg
	nd.initwg.Add(1)
	return nd
}

// SetMaxConcurrentUpdates overrides the default (1000) limit on goroutines
// processing updates concurrently. Must be called before Initialize.
func (dp *NativeDispatcher) SetMaxConcurrentUpdates(n int) {
	dp.updateSem = make(chan struct{}, n)
}

// ConvManager exposes the underlying conversation manager instance.
func (dp *NativeDispatcher) ConvManager() *conv.Manager {
	return dp.conv
}

func defaultErrorHandler(_ *adapter.Context, _ *adapter.Update, err string) error {
	log.Println("An error occured while handling update:", err)
	return ContinueGroups
}

type entities tg.Entities

func (u *entities) short() {
	u.Short = true
	u.Users = nil
	u.Chats = nil
	u.Channels = nil
}

func (dp *NativeDispatcher) Initialize(ctx context.Context, cancel context.CancelFunc, client *telegram.Client, self *tg.User) {
	dp.client = client.API()
	dp.sender = message.NewSender(dp.client)
	dp.self = self
	dp.cancel = cancel
	dp.initwg.Done()
}
