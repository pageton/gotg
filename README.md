<div align="center">
  <img src="./gotgproto.png" width="120px" alt="GoTG Logo">
  
  # GoTG
  
  ### High-Level Telegram MTProto Framework for Go
  
  [![Go Reference](https://pkg.go.dev/badge/github.com/pageton/gotg.svg)](https://pkg.go.dev/github.com/pageton/gotg)
  [![Go Report Card](https://goreportcard.com/badge/github.com/pageton/gotg)](https://goreportcard.com/report/github.com/pageton/gotg)
  [![License: GPL v3](https://img.shields.io/badge/License-GPLv3-blue.svg)](https://www.gnu.org/licenses/gpl-3.0)
  [![Telegram Chat](https://img.shields.io/badge/Chat-Telegram-2CA5E0?logo=telegram&logoColor=white)](https://t.me/gotg_community)
  
  **Production-ready Telegram bot and userbot framework built on [gotd/td](https://github.com/gotd/td)**
  
  [Features](#features) • [Installation](#installation) • [Quick Start](#quick-start) • [Documentation](#documentation) • [Examples](#examples) • [Contributing](#contributing)
  
</div>

---

## About

**GoTG** (Go Telegram) is a comprehensive, high-level framework for building Telegram bots and userbots in Go. Built on top of the robust [gotd/td](https://github.com/gotd/td) MTProto implementation, GoTG provides an intuitive API that abstracts away the complexity of raw Telegram API calls while maintaining full access to underlying functionality.

### Why GoTG?

- **Production-Ready**: Battle-tested with structured logging, error handling, and performance optimizations
- **Developer-Friendly**: Intuitive API with fluent builders, conversation flows, and comprehensive examples
- **Enterprise Features**: Built-in i18n, business API support, and outgoing message tracking
- **High Performance**: Configurable concurrency, optimized string operations, and efficient DB writes
- **Fully Extensible**: Access to raw gotd/td API alongside high-level helpers

> **Note**: This is a comprehensive fork of [celestix/gotgproto](https://github.com/celestix/gotgproto) with **204% more code**, **13 new features**, and **90 commits** of enhancements. See [CHANGELOG.md](./CHANGELOG.md) for detailed comparison.

---

## Features

<details open>
<summary><b>Core Features</b></summary>

- **Session Management**: SQLite, in-memory, and string sessions with seamless import from Pyrogram/Telethon/Gramjs
- **Update Handlers**: 19 handler types covering messages, edits, deletions, reactions, boosts, and more
- **Peer Storage**: Automatic caching of user/chat/channel access hashes
- **Fluent Keyboard API**: Chainable inline and reply keyboard builders
- **Parse Modes**: MarkdownV2 and HTML formatting with entity support
- **Media Handling**: Simplified upload/download with parallel chunk support
- **Middleware System**: Request/response interception and modification
- **Raw API Access**: Direct access to all 1000+ Telegram API methods

</details>

<details>
<summary><b>Advanced Features (Exclusive to GoTG)</b></summary>

### Structured Logging
Full logging infrastructure with per-update loggers, dual console+file output, and configurable levels.

```go
client, _ := gotg.NewClient(apiID, apiHash, gotg.AsBot(token), &gotg.ClientOpts{
    LogConfig: &log.Config{
        MinLevel:  log.LevelDebug,
        Timestamp: true,
        Color:     true,
        Caller:    true,
        LogFile:   "/var/log/bot.log",
    },
})

// In handlers
func handler(u *adapter.Update) error {
    u.Log.Info("processing", "user", u.GetUserChat().ID)
    u.Log.Success("completed") // Green checkmark
    return nil
}
```

### Conversation System
State machine for multi-step user interactions with timeout handling and type filters.

```go
// Register steps
convManager.RegisterStep("reg:name", func(state *conv.State) error {
    name := state.Text()
    state.Set("user_name", name)
    return state.Next("reg:photo", "Send your photo", &conv.NextOpts{
        Filter: conv.Filters.Photo,
        Timeout: 5 * time.Minute,
    })
})

// Start conversation
update.StartConv("reg:name", "What's your name?", &adapter.ConvOpts{
    Reply: true,
    Filter: conv.Filters.Text,
})
```

### Internationalization (i18n)
Support for 142 languages with Fluent (.ftl) and YAML formats, CLDR plural rules.

```go
i18n, _ := i18n.NewI18n(&i18n.Config{
    DefaultLocale: "en",
    LocalesDir: "./locales",
})

text := i18n.Get(userID, "welcome", "name", "Alice")
// en: "Hello, Alice!"
// es: "¡Hola, Alice!"
```

### Business API
Full Telegram Business integration with 5 dedicated handlers and auto-wrapping.

```go
dispatcher.AddHandler(handlers.OnBusinessMessage(func(u *adapter.Update) error {
    return u.Reply("Business reply")
}))
```

### Outgoing Message Tracking
Synthetic updates for sent/edited/deleted messages (compliance, logging, auditing).

```go
gotg.NewClient(apiID, apiHash, gotg.AsBot(token), &gotg.ClientOpts{
    SendOutgoing: true,
})

dispatcher.AddHandler(handlers.OnOutgoing(func(u *adapter.Update) error {
    if u.EffectiveOutgoing.IsSend {
        log.Printf("Sent: %s", u.EffectiveMessage.Text)
    }
    return nil
}))
```

### Debug JSON Dump
Pretty-print any update/struct for debugging (uses bytedance/sonic).

```go
update.Dump("INCOMING") // Clean, formatted JSON output
```

### Performance Optimizations
- **40% faster** string escaping (static lookup table)
- **O(n) string building** (pre-allocated buffers)
- **Bounded goroutines** (configurable semaphore, default 1000)
- **DB write fan-in** (single writer queue)
- **Indexed username lookups**

</details>

---

## Installation

```bash
go get github.com/pageton/gotg
```

**Requirements:**
- Go 1.21 or higher
- Telegram API credentials ([get them here](https://my.telegram.org/apps))

---

## Quick Start

### Bot Example

```go
package main

import (
    "log"
    
    "github.com/pageton/gotg"
    "github.com/pageton/gotg/adapter"
    "github.com/pageton/gotg/dispatcher/handlers"
    "github.com/pageton/gotg/dispatcher/handlers/filters"
)

func main() {
    client, err := gotg.NewClient(
        123456,           // API ID from https://my.telegram.org/apps
        "YOUR_API_HASH",  // API Hash
        gotg.AsBot("YOUR_BOT_TOKEN"),
        &gotg.ClientOpts{
            InMemory: true, // Use in-memory session (no file)
        },
    )
    if err != nil {
        log.Fatalln("failed to start:", err)
    }
    
    dp := client.Dispatcher
    
    // Command handler
    dp.AddHandler(handlers.OnCommand("start", func(u *adapter.Update) error {
        _, err := u.Reply("Hello! I'm alive")
        return err
    }, filters.Private))
    
    // Echo handler
    dp.AddHandler(handlers.OnMessage(func(u *adapter.Update) error {
        _, err := u.Reply(u.Text())
        return err
    }, filters.Message.Text))
    
    log.Println("Bot started!")
    client.Idle()
}
```

### Userbot Example

```go
client, err := gotg.NewClient(
    123456,
    "YOUR_API_HASH",
    gotg.ClientTypePhone("PHONE_NUMBER"),
    &gotg.ClientOpts{
        Session: session.SqlSession(sqlite.Open("userbot.db")),
    },
)
```

---

## Documentation

### Core Concepts

<details>
<summary><b>Handlers & Filters</b></summary>

GoTG provides 19 update handler types:

```go
// Messages
handlers.OnMessage(handler, filters.Message.Text)
handlers.OnEditedMessage(handler, filters.Message.Photo)
handlers.OnDeletedMessage(handler)

// Commands
handlers.OnCommand("help", handler, filters.Private)

// Callbacks & Inline
handlers.OnCallbackQuery(filters.Callback.Data("button_id"), handler)
handlers.OnInlineQuery(handler)
handlers.OnChosenInlineResult(handler)

// Chat Events
handlers.OnChatMemberUpdated(handler)
handlers.OnChatJoinRequest(handler)

// Reactions & Engagement
handlers.OnMessageReaction(handler)
handlers.OnChatBoost(handler)

// Business API
handlers.OnBusinessMessage(handler)
handlers.OnBusinessEditedMessage(handler)
handlers.OnBusinessDeletedMessage(handler)
handlers.OnBusinessConnection(handler)
handlers.OnBusinessCallbackQuery(handler)

// Outgoing Tracking
handlers.OnOutgoing(handler)

// Catch-all
handlers.OnUpdate(handler) // Receives all updates
```

**Common Filters:**

```go
// Message type filters
filters.Message.Text
filters.Message.Photo
filters.Message.Video
filters.Message.Document
filters.Message.Voice
filters.Message.Audio
filters.Message.Sticker
filters.Message.Animation

// Content filters
filters.Message.HasEntities
filters.Message.HasMention
filters.Message.HasHashtag
filters.Message.HasURL
filters.Message.Command
filters.Message.Reply

// Chat type filters
filters.Private
filters.Group
filters.Channel
filters.Supergroup

// Entity filters
filters.Message.From(userID)
filters.Message.Chat(chatID)

// Regex filters
filters.Message.Regex(`^/start (.+)`)
```

</details>

<details>
<summary><b>Sending Messages</b></summary>

```go
// Simple text message
msg, err := update.SendMessage(chatID, "Hello!")

// With formatting
msg, err := update.SendMessage(chatID, "<b>Bold</b> and <i>italic</i>", &adapter.SendOpts{
    ParseMode: adapter.ModeHTML,
})

// With keyboard
kb := keyboard.NewInline().
    Row().URL("Visit", "https://example.com").
    Row().Callback("Click me", "btn_data").
    Build()

msg, err := update.SendMessage(chatID, "Choose:", &adapter.SendOpts{
    ReplyMarkup: kb,
})

// Reply to a message
msg, err := update.Reply("Reply text")

// With media
media := &tg.InputMediaUploadedPhoto{File: uploadedFile}
msg, err := update.SendMedia(media, "Caption", &adapter.SendMediaOpts{
    ParseMode: adapter.ModeMarkdownV2,
})
```

</details>

<details>
<summary><b>Media Upload & Download</b></summary>

**Upload:**

```go
// From local file
f, err := uploader.NewUploader(ctx.Raw).FromPath(ctx, "photo.jpg")
if err != nil {
    return err
}

media := &tg.InputMediaUploadedPhoto{File: f}
msg, err := ctx.SendMedia(chatID, &tg.MessagesSendMediaRequest{
    Media:   media,
    Message: "Check this out!",
})
```

**Download:**

```go
// Download to file
err := msg.DownloadMedia(&adapter.DownloadToPath{Path: "./downloads/file.jpg"})

// Download to memory
var buf bytes.Buffer
err := msg.DownloadMedia(&adapter.DownloadToStream{Writer: &buf})
```

</details>

<details>
<summary><b>Keyboard Builders</b></summary>

**Inline Keyboards:**

```go
kb := keyboard.NewInline().
    Row().
        URL("GitHub", "https://github.com/pageton/gotg").
        URL("Docs", "https://pkg.go.dev/github.com/pageton/gotg").
    Row().
        Callback("Option 1", "opt_1").
        Callback("Option 2", "opt_2").
    Row().
        SwitchInline("Share", "check this out").
    Build()
```

**Reply Keyboards:**

```go
kb := keyboard.NewReply().
    Row().
        Text("Button 1").
        Text("Button 2").
    Row().
        RequestContact("Share Contact").
        RequestLocation("Share Location").
    BuildReply(keyboard.ReplyOptions{
        Resize:    true,
        OneTime:   true,
        Selective: false,
    })
```

</details>

<details>
<summary><b>Session Management</b></summary>

**SQLite Session (persistent):**

```go
import "github.com/glebarez/sqlite"

client, _ := gotg.NewClient(apiID, apiHash, clientType, &gotg.ClientOpts{
    Session: session.SqlSession(sqlite.Open("bot.db")),
})
```

**In-Memory Session (temporary):**

```go
client, _ := gotg.NewClient(apiID, apiHash, clientType, &gotg.ClientOpts{
    InMemory: true,
})
```

**String Session (Pyrogram/Telethon/Gramjs):**

```go
// Pyrogram format
client, _ := gotg.NewClient(apiID, apiHash, clientType, &gotg.ClientOpts{
    Session: session.NewFromString("PYROGRAM_SESSION_STRING", session.Pyrogram),
})

// Telethon format
client, _ := gotg.NewClient(apiID, apiHash, clientType, &gotg.ClientOpts{
    Session: session.NewFromString("TELETHON_SESSION_STRING", session.Telethon),
})
```

</details>

<details>
<summary><b>Error Handling & Middleware</b></summary>

**Custom Error Handler:**

```go
errorHandler := func(ctx *adapter.Context, u *adapter.Update, err string) error {
    log.Printf("Error in update %d: %s", u.MsgID(), err)
    return dispatcher.ContinueGroups // Continue processing
}

dispatcher := dispatcher.NewNativeDispatcher(
    true, true,
    errorHandler,  // Error handler
    nil,           // Panic handler
    peerStorage,
    logger,
    false,
)
```

**Middleware:**

```go
// Logging middleware
dp.AddHandlerToGroup(dispatcher.Handler{
    Handle: func(ctx *adapter.Context, u *adapter.Update) error {
        start := time.Now()
        log.Printf("Processing update %d", u.MsgID())
        
        // Call next handler (implementation depends on your flow)
        
        log.Printf("Completed in %v", time.Since(start))
        return nil
    },
}, -1) // Group -1 runs before all others
```

</details>

---

## Examples

Check the [`examples/`](./examples) directory for complete, runnable examples:

| Example | Description | Features Demonstrated |
|---------|-------------|----------------------|
| [**echo-bot**](./examples/echo-bot) | Simple echo bot | Basic message handling |
| [**business-bot**](./examples/business-bot) | Business API showcase | All 5 business handlers, auto-wrapping |
| [**conversation**](./examples/conversation) | Multi-step registration | State machine, filters, timeouts |
| [**i18n-bot**](./examples/i18n-bot) | Multilingual bot | Fluent FTL, plural rules, language switching |
| [**logging**](./examples/logging) | Structured logging | Per-update loggers, log levels, dual output |
| [**outgoing-updates**](./examples/outgoing-updates) | Outgoing tracking | Sent/edited/deleted message monitoring |
| [**format-example**](./examples/format-example) | Text formatting | MarkdownV2/HTML parsing, entities |
| [**downloader**](./examples/downloader) | Media download | File download with progress |
| [**downloader-v2**](./examples/downloader-v2) | Advanced download | Parallel chunks, resumable downloads |
| [**middleware**](./examples/middleware) | Custom middleware | Request interception, timing |
| [**dispatcher-bot**](./examples/dispatcher-bot) | Handler groups | Priority-based routing |
| [**auth-using-***](./examples/) | Various auth methods | SQLite, string, TData, API sessions |

---

## Configuration

### Client Options

```go
&gotg.ClientOpts{
    // Session Management
    Session:  session.SqlSession(...),  // Session storage
    InMemory: false,                     // Use in-memory session

    // Update Handling
    AutoFetchReply:        true,   // Auto-fetch replied messages
    FetchEntireReplyChain: false,  // Fetch full reply chain
    SendOutgoing:          false,  // Enable outgoing update tracking
    MaxConcurrentUpdates:  1000,   // Goroutine limit (default 1000)
    NoUpdates:             false,  // Disable update processing

    // Logging
    LogConfig: &log.Config{
        MinLevel:  log.LevelInfo,
        Timestamp: true,
        Color:     true,
        Caller:    false,
        FuncName:  false,
        LogFile:   "",  // Optional file output
    },

    // Handlers
    ErrorHandler:  customErrorHandler,
    PanicHandler:  customPanicHandler,
    
    // Middleware
    DispatcherMiddlewares: []dispatcher.Handler{...},
    RunMiddleware:         customMiddleware,

    // Localization
    SystemLangCode: "en",
    ClientLangCode: "en",

    // Authentication
    NoAutoAuth:        false,  // Disable automatic auth
    SendCodeOptions:   &auth.SendCodeOptions{},
    AuthConversator:   customConversator,

    // Network (advanced)
    DC:                2,
    DCList:            customDCList,
    Resolver:          customResolver,
    PublicKeys:        customKeys,
    
    // Timeouts
    MigrationTimeout: 15 * time.Second,
    DialTimeout:      10 * time.Second,
    ExchangeTimeout:  60 * time.Second,
    
    // MTProto
    AckBatchSize:      100,
    AckInterval:       15 * time.Second,
    RetryInterval:     5 * time.Second,
    MaxRetries:        5,
    CompressThreshold: 1024,
}
```

---

## Advanced Usage

### Working with Raw API

For functionality not yet wrapped, use the raw gotd/td API:

```go
// Access raw client
rawClient := ctx.Raw

// Call any MTProto method
history, err := rawClient.MessagesGetHistory(ctx, &tg.MessagesGetHistoryRequest{
    Peer:  peerStorage.GetInputPeerByID(chatID),
    Limit: 10,
})

// Extract entities from update
users := update.Entities.Users
chats := update.Entities.Chats
```

### Custom Handlers

```go
type MyCustomHandler struct {
    // Your fields
}

func (h *MyCustomHandler) CheckUpdate(ctx *adapter.Context, u *adapter.Update) bool {
    // Return true if this handler should process the update
    return u.HasMessage() && u.Text() == "/custom"
}

func (h *MyCustomHandler) HandleUpdate(ctx *adapter.Context, u *adapter.Update) error {
    // Handle the update
    _, err := u.Reply("Custom handler triggered!")
    return err
}

// Register
dp.AddHandler(MyCustomHandler{})
```

---

## Contributing

We welcome contributions! Please see our [Contributing Guidelines](./CONTRIBUTING.md) for details.

**Quick steps:**

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Write/update tests
5. Update documentation
6. Commit your changes (`git commit -m 'feat: add amazing feature'`)
7. Push to the branch (`git push origin feature/amazing-feature`)
8. Open a Pull Request

**Please:**
- Follow existing code style
- Write clear commit messages ([Conventional Commits](https://www.conventionalcommits.org/))
- Add examples for new features
- Update CHANGELOG.md

---

## License

This project is licensed under the **GNU General Public License v3.0** - see the [LICENSE](./LICENSE) file for details.

[![GPLv3](https://www.gnu.org/graphics/gplv3-127x51.png)](https://www.gnu.org/licenses/gpl-3.0.en.html)

---

## Acknowledgments

GoTG is built upon the excellent work of:

- **[celestix/gotgproto](https://github.com/celestix/gotgproto)** — Original framework (core session management, filters, handlers)
- **[gotd/td](https://github.com/gotd/td)** — Robust MTProto implementation
- All contributors to both projects

This fork adds production features (logging, i18n, business API, conversations, performance optimizations) while maintaining full backward compatibility with celestix/gotgproto.

---

## Resources

- **Documentation**: [pkg.go.dev/github.com/pageton/gotg](https://pkg.go.dev/github.com/pageton/gotg)
- **Changelog**: [CHANGELOG.md](./CHANGELOG.md) (detailed feature comparison)
- **Telegram Support**: [@gotg_community](https://t.me/gotg_community)
- **Issue Tracker**: [GitHub Issues](https://github.com/pageton/gotg/issues)
- **Telegram API Docs**: [core.telegram.org/api](https://core.telegram.org/api)
- **MTProto Spec**: [core.telegram.org/mtproto](https://core.telegram.org/mtproto)

---

## Project Status

**Current Version:** `v1.0.0-beta23`

- **Stable Core**: Battle-tested session management and update handling
- **Production Ready**: Used in production bots
- **Beta Status**: Minor API changes possible before v1.0.0 stable
- **Active Development**: Regular updates and improvements

**Stats:**
- 15,304 lines of Go code
- 169 source files
- 3,296 inline documentation comments
- 15 comprehensive examples
- 90 commits since fork

---

<div align="center">
  
  **Star this repo if you find it useful!**
  
</div>
