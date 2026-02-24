package dispatcher

import (
	"context"
	"errors"
	"log"
	"runtime/debug"
	"strings"

	"github.com/gotd/td/tg"
	"github.com/pageton/gotg/adapter"
	"github.com/pageton/gotg/storage"
	"go.uber.org/multierr"
)

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
	case *tg.UpdatesTooLong:
		// Handled by the gap manager (updates.Manager) which tracks
		// pts/qts/seq and calls getDifference with correct state values.
		return nil
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
	dp.updateWg.Add(1)
	dp.updateSem <- struct{}{}
	go func() {
		defer func() { <-dp.updateSem }()
		defer dp.updateWg.Done()
		if err := dp.handleUpdate(ctx, e, update); err != nil {
			if !isControlError(err) {
				log.Println("dispatcher: update handler error:", err)
			}
		}
	}()
	return nil
}

func isSyntheticOutgoing(update tg.UpdateClass) bool {
	u, ok := update.(*tg.UpdateNewMessage)
	if !ok {
		return false
	}
	m, ok := u.Message.(*tg.Message)
	if !ok {
		return false
	}
	return m.Out && u.Pts == 0
}

func isControlError(err error) bool {
	return errors.Is(err, ContinueGroups) || errors.Is(err, EndGroups) || errors.Is(err, SkipCurrentGroup)
}

func (dp *NativeDispatcher) handleUpdate(ctx context.Context, e tg.Entities, update tg.UpdateClass) error {
	c := adapter.NewContext(ctx, dp.client, dp.pStorage, dp.self, dp.sender, &e, dp.setReply, dp.conv, dp.logger, dp.telegramClient)
	c.DefaultParseMode = dp.defaultParseMode
	c.GetDCPool = dp.getDCPool
	if dp.outgoing && !isSyntheticOutgoing(update) {
		c.OnOutgoing = func(ou *adapter.FakeOutgoingUpdate) {
			if ou.Message == nil || ou.Message.Message == nil {
				return
			}
			ou.Message.Out = true
			msg := ou.Message.Message
			dp.pendingOutgoing.Store(msg, ou)
			synth := &tg.UpdateNewMessage{
				Message: msg,
				Pts:     0,
			}
			_ = dp.dispatch(ctx, e, synth) //nolint:errcheck // synthetic outgoing dispatch errors are non-fatal
		}
	}
	u := adapter.GetNewUpdate(c, update)
	if isSyntheticOutgoing(update) {
		if m, ok := update.(*tg.UpdateNewMessage); ok {
			if val, loaded := dp.pendingOutgoing.LoadAndDelete(m.Message); loaded {
				u.EffectiveOutgoing = val.(*adapter.FakeOutgoingUpdate)
			}
		}
	}
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

	dp.handlerGroupsMu.RLock()
	groups := dp.handlerGroups
	dp.handlerGroupsMu.RUnlock()

	for _, group := range groups {
		handlers := dp.handlerMap[group]
		for _, handler := range handlers {
			err = handler.CheckUpdate(c, u)
			switch {
			case err == nil || err == ContinueGroups:
				continue
			case err == EndGroups:
				return err
			case err == SkipCurrentGroup:
				break
			case err == StopClient:
				dp.cancel()
				return nil
			default:
				err = dp.Error(c, u, err.Error())
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

func buildErrorStack(r any) string {
	stack := debug.Stack()

	var sb strings.Builder
	sb.Grow(len(stack) + 64)
	sb.WriteString("panic: ")
	sb.WriteString(interfaceToString(r))
	sb.WriteString("\n")
	sb.Write(stack)
	return sb.String()
}

func interfaceToString(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	if err, ok := v.(error); ok {
		return err.Error()
	}
	return "unknown panic"
}

func (dp *NativeDispatcher) handleUpdateRepliedToMessage(u *adapter.Update, ctx context.Context) {
	if !u.HasMessage() || !dp.setReply {
		return
	}
	msg := u.EffectiveMessage
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
		if ok {
			if !channel.Min {
				p.AddPeer(channel.ID, channel.AccessHash, storage.TypeChannel, channel.Username)
			}
			continue
		}
		c, ok := chat.(*tg.Chat)
		if !ok {
			continue
		}
		p.AddPeer(c.ID, storage.DefaultAccessHash, storage.TypeChat, storage.DefaultUsername)
	}
}
