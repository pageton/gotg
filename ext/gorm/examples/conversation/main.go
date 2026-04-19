package main

import (
	"fmt"
	"log"
	"time"

	"github.com/pageton/gotg"
	"github.com/pageton/gotg/adapter"
	"github.com/pageton/gotg/conv"
	"github.com/pageton/gotg/dispatcher"
	"github.com/pageton/gotg/dispatcher/handlers"
	"github.com/pageton/gotg/dispatcher/handlers/filters"
	gorm "github.com/pageton/gotg/ext/gorm"
	"github.com/pageton/gotg/session"
	"gorm.io/driver/sqlite"
)

func main() {
	adapter, err := gorm.New(sqlite.Open("conv_bot"))
	if err != nil {
		log.Fatalln("failed to create adapter:", err)
	}
	client, err := gotg.NewClient(
		0,                            // APP_ID from https://my.telegram.org/apps
		"",                           // API_HASH from https://my.telegram.org/apps
		gotg.AsBot("BOT_TOKEN_HERE"), // Bot token from @BotFather
		&gotg.ClientOpts{
			Session: session.WithAdapter(adapter),
		},
	)
	if err != nil {
		log.Fatalln("failed to start client:", err)
	}

	dp := client.Dispatcher
	convManager := client.ConvManager

	// 1. Register step handlers for different flows
	convManager.RegisterStep("reg:name", registerName)
	convManager.RegisterStep("reg:photo", registerPhoto)
	convManager.RegisterStep("feedback:msg", feedbackMsg)

	// 2. Add conv middleware with options
	// This handles automatic routing and "cancel" keywords
	dp.AddHandlerToGroup(dispatcher.Conv(dispatcher.ConvOpts{
		Manager:        convManager,
		CancelKeywords: []string{"cancel", "stop"},
		CancelText:     "Conversation stopped by user.",
		CancelReply:    true,
	}), 0)

	// 3. Command handlers
	dp.AddHandler(handlers.OnCommand("start", cmdStart, filters.Private))
	dp.AddHandler(handlers.OnCommand("register", cmdRegister, filters.Private))
	dp.AddHandler(handlers.OnCommand("feedback", cmdFeedback, filters.Private))
	dp.AddHandler(handlers.OnCommand("cancel", cmdCancel, filters.Private))

	fmt.Printf("Conversation bot (@%s) started...\n", client.Self.Username)

	if err = client.Idle(); err != nil {
		log.Fatalln("failed to run client:", err)
	}
}

func cmdStart(u *adapter.Update) error {
	_, err := u.Reply("Welcome! Commands:\n/register - Start registration\n/feedback - Send feedback\n/cancel - Stop any action")
	return err
}

func cmdCancel(u *adapter.Update) error {
	success, wasActive, err := u.EndConv()
	if err != nil {
		return err
	}
	if wasActive && success {
		_, err = u.Reply("Active conv has been terminated.")
		return err
	}
	_, err = u.Reply("No active conv to cancel.")
	return err
}

func cmdRegister(u *adapter.Update) error {
	_, err := u.StartConv("reg:name", "Starting registration! What is your name?", &adapter.ConvOpts{
		Reply:   true,
		Timeout: 5 * time.Minute,
		Filter:  conv.Filters.Text,
	})
	return err
}

func registerName(state *conv.State) error {
	name := state.Text()
	state.Set("user_name", name)
	return state.Next("reg:photo", fmt.Sprintf("Nice to meet you %s! Please send me a photo for your profile.", name), &conv.NextOpts{
		Filter: conv.Filters.Photo,
	})
}

func registerPhoto(state *conv.State) error {
	name := state.GetString("user_name")
	return state.End(fmt.Sprintf("Thank you %s! I've received your photo and your registration is complete.", name), &conv.EndOpts{
		Reply: true,
	})
}

func cmdFeedback(u *adapter.Update) error {
	_, err := u.StartConv("feedback:msg", "Please send your feedback message:", &adapter.ConvOpts{
		Filter: conv.Filters.Text,
	})
	return err
}

func feedbackMsg(state *conv.State) error {
	msg := state.Text()
	log.Printf("Received feedback: %s", msg)
	return state.End("Thank you for your feedback!", &conv.EndOpts{Reply: true})
}
