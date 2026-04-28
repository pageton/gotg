// Package adapter provides update handling, context management, message operations,
// conversation support, and command registry for gotg Telegram bots and userbots.
// It wraps gotd/td types into developer-friendly abstractions with built-in
// logging, i18n, and outgoing message tracking.

package adapter

import (
	"context"
	"fmt"

	"github.com/gotd/td/tg"
)

// BotCmd represents a single bot command with its command string and description.
// Commands are shown in the Telegram client's command menu when users type "/".
//
// Example:
//
//	cmd := BotCmd{
//	    Command:     "start",
//	    Description: "Start the bot",
//	}
type BotCmd struct {
	// Command is the command name without the leading slash (e.g., "start", not "/start").
	// Must be lowercase, 1-32 characters, and contain only letters, digits, and underscores.
	Command string
	// Description is a short description shown in the command menu, 1-256 characters.
	Description string
}

// CommandRegistry builds and registers Telegram bot commands using a fluent API.
// It allows setting different commands for group chats and private chats.
//
// Telegram supports different command sets based on the chat type:
//   - Group commands: shown in groups and supergroups
//   - Private commands: shown in private chats with the bot
//
// Example:
//
//	err := NewCommandRegistry(raw, ctx).
//	    AddGroup(
//	        BotCmd{Command: "help", Description: "Show help"},
//	        BotCmd{Command: "settings", Description: "Bot settings"},
//	    ).
//	    AddPrivate(
//	        BotCmd{Command: "start", Description: "Start the bot"},
//	    ).
//	    WithLangCode("en").
//	    Register()
//	if err != nil {
//	    log.Fatal(err)
//	}
type CommandRegistry struct {
	raw      *tg.Client
	ctx      context.Context
	group    []tg.BotCommand
	private  []tg.BotCommand
	langCode string
}

// NewCommandRegistry creates a new command registry builder.
//
// Parameters:
//   - raw: The raw tg.Client from gotd/td for making API calls.
//   - ctx: The context for the API requests, can be context.Background() or a timeout context.
//
// Example:
//
//	registry := NewCommandRegistry(client.API(), context.Background())
func NewCommandRegistry(raw *tg.Client, ctx context.Context) *CommandRegistry {
	return &CommandRegistry{
		raw: raw,
		ctx: ctx,
	}
}

// WithLangCode sets the language code for the commands.
// Use this to register different commands for different languages.
// An empty string (default) means the commands apply to all languages.
//
// Common language codes: "en", "ar", "ru", "es", "de", "fa", "tr".
//
// Example:
//
//	// Register English commands
//	NewCommandRegistry(raw, ctx).
//	    WithLangCode("en").
//	    AddPrivate(BotCmd{Command: "start", Description: "Start the bot"}).
//	    Register()
//
//	// Register Arabic commands
//	NewCommandRegistry(raw, ctx).
//	    WithLangCode("ar").
//	    AddPrivate(BotCmd{Command: "start", Description: "تشغيل البوت"}).
//	    Register()
func (r *CommandRegistry) WithLangCode(lang string) *CommandRegistry {
	r.langCode = lang
	return r
}

// AddGroup adds commands for group and supergroup chats.
// These commands appear in the command menu when users type "/" in groups.
//
// Group commands are typically for features that make sense in a group context:
//   - Games and interactive features
//   - Group management commands
//   - Commands that affect the whole group
//
// Example:
//
//	NewCommandRegistry(raw, ctx).
//	    AddGroup(
//	        BotCmd{Command: "startmirror", Description: "بدء لعبة المرآة"},
//	        BotCmd{Command: "startcity", Description: "بدء لعبة المدينة الفاسدة"},
//	        BotCmd{Command: "roulette", Description: "بدء لعبة الروليت"},
//	        BotCmd{Command: "players", Description: "عرض لاعبي اللعبة الحالية"},
//	    ).
//	    Register()
func (r *CommandRegistry) AddGroup(cmds ...BotCmd) *CommandRegistry {
	for _, c := range cmds {
		r.group = append(r.group, tg.BotCommand{
			Command:     c.Command,
			Description: c.Description,
		})
	}
	return r
}

// AddPrivate adds commands for private chats with the bot.
// These commands appear in the command menu when users type "/" in direct messages.
//
// Private commands are typically for:
//   - Bot onboarding (/start)
//   - User-specific settings
//   - Personal features
//
// Example:
//
//	NewCommandRegistry(raw, ctx).
//	    AddPrivate(
//	        BotCmd{Command: "start", Description: "تشغيل البوت"},
//	        BotCmd{Command: "help", Description: "المساعدة"},
//	        BotCmd{Command: "settings", Description: "الإعدادات"},
//	    ).
//	    Register()
func (r *CommandRegistry) AddPrivate(cmds ...BotCmd) *CommandRegistry {
	for _, c := range cmds {
		r.private = append(r.private, tg.BotCommand{
			Command:     c.Command,
			Description: c.Description,
		})
	}
	return r
}

// AddGroupMap adds commands from a map for group chats.
// This is convenient when commands are defined in a configuration or i18n structure.
//
// Example:
//
//	groupCmds := map[string]string{
//	    "startmirror": "بدء لعبة المرآة",
//	    "startcity":   "بدء لعبة المدينة الفاسدة",
//	    "roulette":    "بدء لعبة الروليت",
//	}
//
//	NewCommandRegistry(raw, ctx).
//	    AddGroupMap(groupCmds).
//	    Register()
func (r *CommandRegistry) AddGroupMap(cmds map[string]string) *CommandRegistry {
	for cmd, desc := range cmds {
		r.group = append(r.group, tg.BotCommand{
			Command:     cmd,
			Description: desc,
		})
	}
	return r
}

// AddPrivateMap adds commands from a map for private chats.
// This is convenient when commands are defined in a configuration or i18n structure.
//
// Example:
//
//	privateCmds := map[string]string{
//	    "start":    "تشغيل البوت",
//	    "help":     "المساعدة",
//	    "settings": "الإعدادات",
//	}
//
//	NewCommandRegistry(raw, ctx).
//	    AddPrivateMap(privateCmds).
//	    Register()
func (r *CommandRegistry) AddPrivateMap(cmds map[string]string) *CommandRegistry {
	for cmd, desc := range cmds {
		r.private = append(r.private, tg.BotCommand{
			Command:     cmd,
			Description: desc,
		})
	}
	return r
}

// Register sends all accumulated commands to Telegram.
// Call this after adding all commands with AddGroup/AddPrivate methods.
//
// Register returns an error if any API call fails. On success, the commands
// will immediately appear in the Telegram client's command menu.
//
// Errors can occur if:
//   - The bot token is invalid
//   - Network issues prevent the API call
//   - Command format is invalid (e.g., uppercase letters, special characters)
//
// Example:
//
//	err := NewCommandRegistry(raw, ctx).
//	    AddGroup(BotCmd{Command: "help", Description: "Show help"}).
//	    AddPrivate(BotCmd{Command: "start", Description: "Start the bot"}).
//	    Register()
//	if err != nil {
//	    log.Printf("failed to register commands: %v", err)
//	}
func (r *CommandRegistry) Register() error {
	if len(r.group) > 0 {
		_, err := r.raw.BotsSetBotCommands(r.ctx, &tg.BotsSetBotCommandsRequest{
			Scope:    &tg.BotCommandScopeChats{},
			LangCode: r.langCode,
			Commands: r.group,
		})
		if err != nil {
			return fmt.Errorf("register group commands: %w", err)
		}
	}

	if len(r.private) > 0 {
		_, err := r.raw.BotsSetBotCommands(r.ctx, &tg.BotsSetBotCommandsRequest{
			Scope:    &tg.BotCommandScopeUsers{},
			LangCode: r.langCode,
			Commands: r.private,
		})
		if err != nil {
			return fmt.Errorf("register private commands: %w", err)
		}
	}

	return nil
}

// MustRegister registers commands and panics on error.
// Use this when you're confident registration should never fail
// (e.g., during initialization with hardcoded commands).
//
// For production code, prefer Register() and handle the error gracefully.
//
// Example:
//
//	func main() {
//	    // ... setup client ...
//
//	    NewCommandRegistry(raw, ctx).
//	        AddGroup(
//	            BotCmd{Command: "help", Description: "Show help"},
//	        ).
//	        AddPrivate(
//	            BotCmd{Command: "start", Description: "Start the bot"},
//	        ).
//	        MustRegister() // Panics if registration fails
//
//	    client.Idle()
//	}
func (r *CommandRegistry) MustRegister() {
	if err := r.Register(); err != nil {
		panic(err)
	}
}

// Clear removes all registered commands for both group and private chats.
// This is useful for resetting the bot's command menu or before
// registering a new set of commands.
//
// Example:
//
//	registry := NewCommandRegistry(raw, ctx)
//
//	// Remove all existing commands
//	if err := registry.Clear(); err != nil {
//	    log.Printf("failed to clear commands: %v", err)
//	}
//
//	// Register new commands
//	registry.
//	    AddPrivate(BotCmd{Command: "start", Description: "Start fresh"}).
//	    Register()
func (r *CommandRegistry) Clear() error {
	_, err := r.raw.BotsSetBotCommands(r.ctx, &tg.BotsSetBotCommandsRequest{
		Scope:    &tg.BotCommandScopeChats{},
		LangCode: r.langCode,
		Commands: nil,
	})
	if err != nil {
		return fmt.Errorf("clear group commands: %w", err)
	}

	_, err = r.raw.BotsSetBotCommands(r.ctx, &tg.BotsSetBotCommandsRequest{
		Scope:    &tg.BotCommandScopeUsers{},
		LangCode: r.langCode,
		Commands: nil,
	})
	if err != nil {
		return fmt.Errorf("clear private commands: %w", err)
	}

	return nil
}

// --- Context Methods ---

// RegisterBotCommands creates a new CommandRegistry from the Context.
// This is a convenience method to register bot commands directly from handlers.
//
// The Context provides both the raw tg.Client and the context.Context needed
// for the registry, so you don't need to pass them separately.
//
// Example:
//
//	func startCmd(u *adapter.Update) error {
//	    // Register commands from within a handler
//	    err := u.Ctx.RegisterBotCommands().
//	        AddPrivate(
//	            adapter.BotCmd{Command: "start", Description: "Start the bot"},
//	            adapter.BotCmd{Command: "help", Description: "Show help"},
//	        ).
//	        AddGroup(
//	            adapter.BotCmd{Command: "play", Description: "Start a game"},
//	        ).
//	        Register()
//	    if err != nil {
//	        u.Log.Error("failed to register commands", "error", err)
//	    }
//	    return u.Reply("Commands registered!")
//	}
func (ctx *Context) RegisterBotCommands() *CommandRegistry {
	return NewCommandRegistry(ctx.Raw, ctx.Context)
}

// SetBotCommands is a convenience method to quickly set bot commands.
// Use this for simple cases where you just want to register commands
// without chaining multiple builder methods.
//
// For more control, use RegisterBotCommands() instead.
//
// Example:
//
//	err := ctx.SetBotCommands(
//	    []adapter.BotCmd{
//	        {Command: "play", Description: "Start a game"},
//	    },
//	    []adapter.BotCmd{
//	        {Command: "start", Description: "Start the bot"},
//	    },
//	)
func (ctx *Context) SetBotCommands(group, private []BotCmd) error {
	return ctx.RegisterBotCommands().
		AddGroup(group...).
		AddPrivate(private...).
		Register()
}
