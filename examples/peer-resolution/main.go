// Package main demonstrates GoTG's automatic peer resolution.
//
// GoTG resolves peers automatically in two ways:
//
//  1. Implicit resolution — pass a chatID to SendMessage/SendMedia and the
//     framework resolves it to an InputPeer via a 3-tier lookup
//     (cache → DB → RPC). You never construct InputPeer yourself.
//
//  2. Explicit resolution — call ResolvePeerToID with a human-readable
//     identifier (@username, +phone, or numeric ID string) to get a chatID,
//     then pass it to any send method.
//
// This example shows both patterns.
package main

import (
	"fmt"
	"log"

	"github.com/pageton/gotg"
	"github.com/pageton/gotg/adapter"
	"github.com/pageton/gotg/dispatcher"
	"github.com/pageton/gotg/dispatcher/handlers"
	"github.com/pageton/gotg/dispatcher/handlers/filters"
	"github.com/pageton/gotg/session"
)

func main() {
	client, err := gotg.NewClient(
		123456,          // AppID from https://my.telegram.org/apps
		"API_HASH_HERE", // ApiHash from https://my.telegram.org/apps
		gotg.AsBot("BOT_TOKEN_HERE"),
		&gotg.ClientOpts{
			InMemory: true,
			Session:  session.SimpleSession(),
			// Pre-load peers from all dialogs so username/phone lookups work.
			// Without this, only peers seen in incoming updates are cached.
			PeersFromDialogs: true,
		},
	)
	if err != nil {
		log.Fatalln("failed to start client:", err)
	}

	dp := client.Dispatcher

	dp.AddHandler(handlers.OnCommand("start", startCmd))
	dp.AddHandler(handlers.OnCommand("send", sendCmd))
	dp.AddHandler(handlers.OnCommand("resolve", resolveCmd))
	dp.AddHandler(handlers.OnMessage(echo, filters.Message.Text))

	log.Printf("Bot @%s started!", client.Self.Username)

	if err := client.Idle(); err != nil {
		log.Fatalln("failed to idle:", err)
	}
}

// startCmd demonstrates the simplest peer resolution: no peer needed at all.
//
// Reply() automatically resolves the current chat from the update context.
// You never specify a chatID or InputPeer — the framework handles it.
func startCmd(u *adapter.Update) error {
	_, err := u.Reply("Hello! I demonstrate automatic peer resolution.\n\n" +
		"Commands:\n" +
		"/send @username — send a message to any user/channel by username\n" +
		"/resolve @username — resolve a username to a chat ID\n" +
		"Any other text — I'll echo it back (implicit resolution)")
	return err
}

// sendCmd demonstrates explicit peer resolution using ResolvePeerToID.
//
// ResolvePeerToID accepts three identifier formats:
//   - "@username" or "username" → resolved via username index → DB → RPC
//   - "+1234567890"            → resolved via phone index → DB → RPC
//   - "-1001234567890"         → parsed as numeric ID
//
// Once you have the chatID, pass it to any send method — SendMessage,
// SendMedia, etc. The send method then resolves chatID → InputPeer
// automatically via the same 3-tier lookup.
func sendCmd(u *adapter.Update) error {
	args := u.Args()
	if len(args) < 2 {
		_, _ = u.Reply("Usage: /send @username <message>")
		return dispatcher.EndGroups
	}

	identifier := args[1]
	text := "Hello from GoTG!"
	if len(args) > 2 {
		text = args[2]
	}

	// Step 1: Resolve the human-readable identifier to a chatID.
	chatID, err := u.ResolvePeerToID(identifier)
	if err != nil {
		_, _ = u.Reply(fmt.Sprintf("Could not resolve %q: %v", identifier, err))
		return dispatcher.EndGroups
	}

	// Step 2: Send the message. SendMessage resolves chatID → InputPeer
	// internally — you never construct InputPeerUser/InputPeerChannel yourself.
	msg, err := u.SendMessage(chatID, text)
	if err != nil {
		_, _ = u.Reply(fmt.Sprintf("Failed to send: %v", err))
		return dispatcher.EndGroups
	}

	_, _ = u.Reply(fmt.Sprintf("Message sent to %s (chatID=%d, msgID=%d)", identifier, chatID, msg.ID))
	return dispatcher.EndGroups
}

// resolveCmd demonstrates resolving a peer and inspecting its metadata.
func resolveCmd(u *adapter.Update) error {
	args := u.Args()
	if len(args) < 2 {
		_, _ = u.Reply("Usage: /resolve @username  OR  /resolve +1234567890")
		return dispatcher.EndGroups
	}

	chatID, err := u.ResolvePeerToID(args[1])
	if err != nil {
		_, _ = u.Reply(fmt.Sprintf("Could not resolve %q: %v", args[1], err))
		return dispatcher.EndGroups
	}

	// Once resolved, you can use chatID with any Context method.
	peer := u.Ctx.ResolvePeerByID(chatID)

	info := fmt.Sprintf("Resolved %q → chatID=%d", args[1], chatID)
	if peer != nil && peer.Username != "" {
		info += fmt.Sprintf(" (@%s)", peer.Username)
	}
	if peer != nil && peer.PhoneNumber != "" {
		info += fmt.Sprintf(" (phone=%s)", peer.PhoneNumber)
	}

	_, _ = u.Reply(info)
	return dispatcher.EndGroups
}

// echo demonstrates the zero-chatID pattern for implicit resolution.
//
// SendMessage(0, text) means "use the current update's chat".
// The framework extracts the chat from the incoming update automatically.
// This is equivalent to u.Reply(text) but without the reply-to reference.
func echo(u *adapter.Update) error {
	// chatID=0 → framework uses u.ChatID() from the current update.
	_, err := u.SendMessage(0, u.Text())
	return err
}
