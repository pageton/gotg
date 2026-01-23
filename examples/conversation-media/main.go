package main

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gotd/td/tg"
	"github.com/pageton/gotg"
	"github.com/pageton/gotg/adapter"
	"github.com/pageton/gotg/conversation"
	"github.com/pageton/gotg/dispatcher/handlers"
	gotgErrors "github.com/pageton/gotg/errors"
	"github.com/pageton/gotg/session"
	"github.com/pageton/gotg/types"
	"gorm.io/driver/sqlite"
)

func main() {
	client, err := gotg.NewClient(
		// Get AppID from https://my.telegram.org/apps
		123456,
		// Get ApiHash from https://my.telegram.org/apps
		"API_HASH_HERE",
		gotg.AsBot("BOT_TOKEN_HERE"),
		&gotg.ClientOpts{
			Session: session.SqlSession(sqlite.Open("conversation_bot.db")),
		},
	)
	if err != nil {
		log.Fatalf("failed to start client: %v", err)
	}

	dp := client.Dispatcher
	dp.AddHandler(handlers.OnCommand("share", shareStatusHandler))

	log.Printf("Media conversation bot ready as @%s", client.Self.Username)
	client.Idle()
}

func shareStatusHandler(u *adapter.Update) error {
	prompt := "Send me a status update — text or a photo works. Type 'cancel' to stop."
	resp, err := askTextOrPhoto(u, prompt)
	if err != nil {
		if errors.Is(err, gotgErrors.ErrConversationCancelled) {
			return nil
		}
		return err
	}

	if photo := extractPhoto(resp); photo != nil {
		_, _ = u.Reply(fmt.Sprintf("📸 Nice photo! ID: %d", photo.ID))
		return nil
	}

	_, _ = u.Reply(fmt.Sprintf("📝 Got your update: %s", resp.Text))
	return nil
}

func askTextOrPhoto(u *adapter.Update, prompt string) (*types.Message, error) {
	resp, err := u.Ask(
		prompt,
		adapter.AskWithTimeout(1*time.Minute),
		adapter.AskWithFilter(acceptTextOrPhoto),
	)
	if err != nil {
		return nil, handleAskError(u, err)
	}
	if isCancel(resp.Text) {
		_, _ = u.Reply("❌ Conversation cancelled.")
		return nil, conversation.ErrCancelled
	}
	return resp, nil
}

func acceptTextOrPhoto(msg *types.Message) bool {
	if msg == nil || msg.Message == nil {
		return false
	}
	if strings.TrimSpace(msg.Text) != "" {
		return true
	}
	_, ok := msg.Media.(*tg.MessageMediaPhoto)
	return ok
}

func extractPhoto(msg *types.Message) *tg.Photo {
	media, ok := msg.Media.(*tg.MessageMediaPhoto)
	if !ok {
		return nil
	}
	photo, _ := media.Photo.(*tg.Photo)
	return photo
}

func handleAskError(u *adapter.Update, err error) error {
	switch {
	case errors.Is(err, conversation.ErrTimeout):
		_, _ = u.Reply("⌛️ Timed out waiting for your reply. Please try /share again.")
	default:
		_, _ = u.Reply(fmt.Sprintf("❌ Conversation failed: %v", err))
	}
	return err
}

func isCancel(text string) bool {
	return strings.EqualFold(strings.TrimSpace(text), "cancel")
}
