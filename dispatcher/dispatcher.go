package dispatcher

import (
	"context"
	"errors"
	"log"
	"runtime/debug"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/tg"
	"github.com/pageton/gotg/adapter"
	"github.com/pageton/gotg/conv"
	"github.com/pageton/gotg/storage"
	"go.uber.org/multierr"
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
	conv                *conv.Manager
	// Panic handles all the panics that occur during handler execution.
	Panic PanicHandler
	// Error handles all the unknown errors which are returned by handler callback functions.
	Error ErrorHandler
	// handlerMap is used for internal functionality of NativeDispatcher.
	handlerMap map[int][]Handler
	// handlerGroups is used for internal functionality of NativeDispatcher.
	handlerGroups []int
	// handlerGroupsMu protects handlerGroups and handlerMap during registration
	handlerGroupsMu sync.RWMutex
	pStorage        *storage.PeerStorage
	initwg          sync.WaitGroup
	nextGroup       atomic.Int64
}

type PanicHandler func(*adapter.Context, *adapter.Update, string)
type ErrorHandler func(*adapter.Context, *adapter.Update, string) error

// MakeDispatcher creates new custom dispatcher which process and handles incoming updates.
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
func NewNativeDispatcher(setReply bool, setEntireReplyChain bool, eHandler ErrorHandler, pHandler PanicHandler, p *storage.PeerStorage) *NativeDispatcher {
	if eHandler == nil {
		eHandler = defaultErrorHandler
	}
	// Pre-allocate handlerGroups with reasonable capacity to avoid early reallocations
	nd := &NativeDispatcher{
		pStorage:            p,
		handlerMap:          make(map[int][]Handler),
		handlerGroups:       make([]int, 0, 8), // Pre-allocate for 8 groups
		setReply:            setReply,
		setEntireReplyChain: setEntireReplyChain,
		Error:               eHandler,
		Panic:               pHandler,
	}
	nd.conv = conv.NewManager(p, 1*time.Minute)
	nd.initwg.Add(1)
	return nd
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

// Optimized: Use nil maps instead of empty maps to avoid allocations
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

// Handle function handles all the incoming updates, map entities and dispatches updates for further handling.
func (dp *NativeDispatcher) Handle(ctx context.Context, updates tg.UpdatesClass) error {
	dp.initwg.Wait()
	var (
		e    entities
		upds []tg.UpdateClass
	)
	switch u := updates.(type) {
	case *tg.Updates:
		upds = u.Updates
		e.Users = u.MapUsers().NotEmptyToMap()
		chats := u.MapChats()
		e.Chats = chats.ChatToMap()
		e.Channels = chats.ChannelToMap()
		saveUsersPeers(u.Users, dp.pStorage)
		saveChatsPeers(u.Chats, dp.pStorage)
	case *tg.UpdatesCombined:
		upds = u.Updates
		e.Users = u.MapUsers().NotEmptyToMap()
		chats := u.MapChats()
		e.Chats = chats.ChatToMap()
		e.Channels = chats.ChannelToMap()
		saveUsersPeers(u.Users, dp.pStorage)
		saveChatsPeers(u.Chats, dp.pStorage)
	case *tg.UpdateShort:
		upds = []tg.UpdateClass{u.Update}
		e.short()
	default:
		return nil
	}

	var err error
	for _, update := range upds {
		multierr.AppendInto(&err, dp.dispatch(ctx, tg.Entities(e), update))
	}
	return err
}

func (dp *NativeDispatcher) dispatch(ctx context.Context, e tg.Entities, update tg.UpdateClass) error {
	if update == nil {
		return nil
	}
	go func() {
		if err := dp.handleUpdate(ctx, e, update); err != nil {
			// Optimized: Single error comparison using switch
			if !isControlError(err) {
				log.Println("dispatcher: update handler error:", err)
			}
		}
	}()
	return nil
}

// isControlError checks if error is a control flow error (not a real error)
// Optimized: Single function for all control error checks
func isControlError(err error) bool {
	return errors.Is(err, ContinueGroups) || errors.Is(err, EndGroups) || errors.Is(err, SkipCurrentGroup)
}

func (dp *NativeDispatcher) handleUpdate(ctx context.Context, e tg.Entities, update tg.UpdateClass) error {
	c := adapter.NewContext(ctx, dp.client, dp.pStorage, dp.self, dp.sender, &e, dp.setReply, dp.conv)
	u := adapter.GetNewUpdate(c, update)
	dp.handleUpdateRepliedToMessage(u, ctx)
	var err error
	defer func() {
		if r := recover(); r != nil {
			errorStack := buildErrorStack(r)
			if dp.Panic != nil {
				dp.Panic(c, u, errorStack)
				return
			} else {
				log.Println(errorStack)
			}
		}
	}()

	// Optimized: Read handler groups once per update (read lock)
	dp.handlerGroupsMu.RLock()
	groups := make([]int, len(dp.handlerGroups))
	copy(groups, dp.handlerGroups)
	dp.handlerGroupsMu.RUnlock()

	for _, group := range groups {
		handlers := dp.handlerMap[group]
		for _, handler := range handlers {
			err = handler.CheckUpdate(c, u)
			// Optimized: Use direct error comparison instead of errors.Is for known sentinel errors
			if err == nil || err == ContinueGroups {
				continue
			} else if err == EndGroups {
				return err
			} else if err == SkipCurrentGroup {
				break
			} else if err == StopClient {
				dp.cancel()
				return nil
			} else {
				err = dp.Error(c, u, err.Error())
				// Optimized: Direct comparison for sentinel errors
				switch err {
				case ContinueGroups:
					continue
				case EndGroups:
					return err
				case SkipCurrentGroup:
					break
				}
			}
		}
	}
	return err
}

// buildErrorStack builds error stack string efficiently using strings.Builder
func buildErrorStack(r any) string {
	stack := debug.Stack()

	// Optimized: Use strings.Builder instead of concatenation
	var sb strings.Builder
	// Pre-allocate capacity based on stack size plus panic message
	sb.Grow(len(stack) + 64)
	sb.WriteString("panic: ")
	sb.WriteString(interfaceToString(r))
	sb.WriteString("\n")
	sb.Write(stack)
	return sb.String()
}

// interfaceToString converts interface{} to string efficiently
func interfaceToString(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	if err, ok := v.(error); ok {
		return err.Error()
	}
	// Fallback
	return "unknown panic"
}

func (dp *NativeDispatcher) handleUpdateRepliedToMessage(u *adapter.Update, ctx context.Context) {
	msg := u.EffectiveMessage
	if msg == nil || !dp.setReply {
		return
	}
	for {
		if msg.Message.ReplyTo == nil {
			return
		}

		_ = msg.SetRepliedToMessage(ctx, dp.client, dp.pStorage)
		if !dp.setEntireReplyChain {
			return
		}
		msg = msg.ReplyToMessage
	}
}

func saveUsersPeers(u tg.UserClassArray, p *storage.PeerStorage) {
	for _, user := range u {
		c, ok := user.AsNotEmpty()
		if !ok || c.Min {
			continue
		}
		p.AddPeer(c.ID, c.AccessHash, storage.TypeUser, c.Username)
	}
}

func saveChatsPeers(u tg.ChatClassArray, p *storage.PeerStorage) {
	for _, chat := range u {
		channel, ok := chat.(*tg.Channel)
		if ok && !channel.Min {
			p.AddPeer(channel.ID, channel.AccessHash, storage.TypeChannel, channel.Username)
			continue
		}
		chat, ok := chat.(*tg.Chat)
		if !ok {
			continue
		}
		p.AddPeer(chat.ID, storage.DefaultAccessHash, storage.TypeChat, storage.DefaultUsername)
	}
}
