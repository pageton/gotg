package main

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/pageton/gotg"
	"github.com/pageton/gotg/adapter"
	"github.com/pageton/gotg/conversation"
	"github.com/pageton/gotg/dispatcher/handlers"
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
	dp.AddHandlers(
		handlers.OnCommand("hello", helloHandler),
		handlers.OnCommand("survey", startSurveyHandler),
	)

	log.Printf("Survey bot ready as @%s", client.Self.Username)
	client.Idle()
}

func helloHandler(u *adapter.Update) error {
	resp, err := u.Ask(
		"👋 How are you doing today?",
		adapter.AskWithTimeout(1*time.Minute),
	)
	if err != nil {
		return handleAskError(u, err)
	}
	_, _ = u.Reply(fmt.Sprintf("Thanks for sharing: %s", resp.Text))
	return nil
}

func startSurveyHandler(u *adapter.Update) error {
	_, _ = u.Reply("📋 Starting survey – answer here or send 'cancel' to stop.")
	name, err := askQuestion(u, "👋 What's your name?", "name")
	if err != nil {
		if errors.Is(err, conversation.ErrCancelled) {
			return nil
		}
		return err
	}
	age, err := askQuestion(u, fmt.Sprintf("Great to meet you, %s! How old are you?", name.Text), "age")
	if err != nil {
		if errors.Is(err, conversation.ErrCancelled) {
			return nil
		}
		return err
	}
	hobby, err := askQuestion(u, "And finally, what's your favorite hobby?", "hobby")
	if err != nil {
		if errors.Is(err, conversation.ErrCancelled) {
			return nil
		}
		return err
	}

	_, _ = u.Reply(fmt.Sprintf(
		"✅ Survey complete!\nName: %s\nAge: %s\nHobby: %s",
		name.Text,
		age.Text,
		hobby.Text,
	))
	return nil
}

func askWithCancel(u *adapter.Update, prompt string, opts ...adapter.AskOption) (*types.Message, error) {
	msg, err := u.Ask(prompt, opts...)
	if err != nil {
		return nil, handleAskError(u, err)
	}
	if isCancel(msg.Text) {
		_, _ = u.Reply("❌ Conversation cancelled.")
		return nil, conversation.ErrCancelled
	}
	return msg, nil
}

func askQuestion(u *adapter.Update, prompt, step string) (*types.Message, error) {
	return askWithCancel(
		u,
		prompt,
		adapter.AskWithStep(step),
		adapter.AskWithTimeout(45*time.Second),
	)
}

func handleAskError(u *adapter.Update, err error) error {
	switch {
	case errors.Is(err, conversation.ErrTimeout):
		_, _ = u.Reply("⌛️ Timed out waiting for your reply. Send /survey to try again.")
	default:
		_, _ = u.Reply(fmt.Sprintf("❌ Conversation failed: %v", err))
	}
	return err
}

func isCancel(text string) bool {
	return strings.EqualFold(strings.TrimSpace(text), "cancel")
}
