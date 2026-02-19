// Package main demonstrates how to register bot commands using CommandRegistry.
// Bot commands appear in Telegram's command menu when users type "/".
package main

import (
	"context"
	"log"

	"github.com/pageton/gotg"
	"github.com/pageton/gotg/adapter"
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
		},
	)
	if err != nil {
		log.Fatalln("failed to start client:", err)
	}

	// Register bot commands using the builder pattern.
	// Commands are split into group commands (for groups/supergroups)
	// and private commands (for direct messages with the bot).
	err = adapter.NewCommandRegistry(client.API(), context.Background()).
		// Group commands - games and group features
		AddGroup(
			adapter.BotCmd{Command: "play", Description: "Start a game"},
			adapter.BotCmd{Command: "stop", Description: "Stop current game"},
			adapter.BotCmd{Command: "players", Description: "Show current players"},
			adapter.BotCmd{Command: "help", Description: "Show help"},
		).
		// Private commands - bot control and settings
		AddPrivate(
			adapter.BotCmd{Command: "start", Description: "Start the bot"},
			adapter.BotCmd{Command: "settings", Description: "Bot settings"},
			adapter.BotCmd{Command: "help", Description: "Show help"},
		).
		Register()
	if err != nil {
		log.Println("failed to register commands:", err)
	} else {
		log.Println("Bot commands registered successfully!")
	}

	dp := client.Dispatcher

	// Command handlers
	dp.AddHandler(handlers.OnCommand("start", startCmd, filters.Private))
	dp.AddHandler(handlers.OnCommand("help", helpCmd))
	dp.AddHandler(handlers.OnCommand("play", playCmd, filters.Group))
	dp.AddHandler(handlers.OnCommand("stop", stopCmd, filters.Group))
	dp.AddHandler(handlers.OnCommand("players", playersCmd, filters.Group))
	dp.AddHandler(handlers.OnCommand("settings", settingsCmd, filters.Private))

	log.Printf("Bot @%s started!", client.Self.Username)

	if err := client.Idle(); err != nil {
		log.Fatalln("failed to idle:", err)
	}
}

func startCmd(u *adapter.Update) error {
	_, err := u.Reply("Welcome! I'm your game bot. Use /help to see available commands.")
	return err
}

func helpCmd(u *adapter.Update) error {
	helpText := `🎮 **Game Bot Commands**

**Group Commands:**
/play - Start a game
/stop - Stop current game
/players - Show current players

**Private Commands:**
/start - Start the bot
/settings - Bot settings`
	_, err := u.Reply(helpText)
	return err
}

func playCmd(u *adapter.Update) error {
	_, err := u.Reply("🎮 Game started! Use /stop to end the game.")
	return err
}

func stopCmd(u *adapter.Update) error {
	_, err := u.Reply("🛑 Game stopped!")
	return err
}

func playersCmd(u *adapter.Update) error {
	_, err := u.Reply("👥 No players currently in the game.")
	return err
}

func settingsCmd(u *adapter.Update) error {
	_, err := u.Reply("⚙️ Settings: Configure your preferences here.")
	return err
}

// Alternative: Using map-based command registration
// This is useful for i18n or configuration-driven setups.
func registerCommandsFromMap(client *gotg.Client) error {
	groupCommands := map[string]string{
		"play":    "Start a game",
		"stop":    "Stop current game",
		"players": "Show current players",
		"help":    "Show help",
	}

	privateCommands := map[string]string{
		"start":    "Start the bot",
		"settings": "Bot settings",
		"help":     "Show help",
	}

	return adapter.NewCommandRegistry(client.API(), context.Background()).
		AddGroupMap(groupCommands).
		AddPrivateMap(privateCommands).
		Register()
}

// Alternative: Multi-language command registration
// Register different descriptions for different languages.
func registerMultiLanguageCommands(client *gotg.Client) error {
	// Default commands (English)
	err := adapter.NewCommandRegistry(client.API(), context.Background()).
		WithLangCode("").
		AddPrivate(
			adapter.BotCmd{Command: "start", Description: "Start the bot"},
			adapter.BotCmd{Command: "help", Description: "Show help"},
		).
		Register()
	if err != nil {
		return err
	}

	// Arabic commands
	return adapter.NewCommandRegistry(client.API(), context.Background()).
		WithLangCode("ar").
		AddPrivate(
			adapter.BotCmd{Command: "start", Description: "تشغيل البوت"},
			adapter.BotCmd{Command: "help", Description: "المساعدة"},
		).
		Register()
}

// Alternative: Register commands from within a handler using Context
func registerFromHandler(u *adapter.Update) error {
	err := u.Ctx.RegisterBotCommands().
		AddGroup(
			adapter.BotCmd{Command: "play", Description: "Start a game"},
			adapter.BotCmd{Command: "stop", Description: "Stop current game"},
		).
		AddPrivate(
			adapter.BotCmd{Command: "start", Description: "Start the bot"},
		).
		Register()
	if err != nil {
		_, _ = u.Reply("Failed to register commands: " + err.Error())
		return err
	}
	_, err = u.Reply("Commands registered successfully!")
	return err
}

// Alternative: Quick registration using SetBotCommands
func quickRegister(u *adapter.Update) error {
	return u.Ctx.SetBotCommands(
		[]adapter.BotCmd{
			{Command: "play", Description: "Start a game"},
			{Command: "stop", Description: "Stop current game"},
		},
		[]adapter.BotCmd{
			{Command: "start", Description: "Start the bot"},
			{Command: "help", Description: "Show help"},
		},
	)
}
