package main

import (
	"embed"
	"fmt"
	"log"

	"github.com/pageton/gotg"
	"github.com/pageton/gotg/adapter"
	"github.com/pageton/gotg/dispatcher"
	"github.com/pageton/gotg/dispatcher/handlers"
	"github.com/pageton/gotg/dispatcher/handlers/filters"
	"github.com/pageton/gotg/i18n"
	"github.com/pageton/gotg/session"
	"golang.org/x/text/language"
	"gorm.io/driver/sqlite"
)

//go:embed locales/*.yaml
var localeFS embed.FS

// Global translator for debugging
var translator *i18n.Translator

func main() {
	// Create translator with embedded locale files
	translator = i18n.NewTranslator(&i18n.LocaleConfig{
		DefaultLang:    language.English,
		Format:         i18n.FormatYAML,
		EmbedFS:        localeFS,
		LocaleDir:      "locales",
		SupportedLangs: []language.Tag{language.English, language.Spanish},
	})

	client, err := gotg.NewClient(
		123456,                       // APP_ID - Replace with your Telegram App ID
		"API_HASH_HERE",              // APP_HASH - Replace with your Telegram App Hash
		gotg.AsBot("BOT_TOKEN_HERE"), // BOT_TOKEN - Replace with your bot token
		&gotg.ClientOpts{
			Session: session.SqlSession(sqlite.Open("testbot")),
			DispatcherMiddlewares: []dispatcher.Handler{
				dispatcher.I18nMiddleware(&dispatcher.I18nConfig{
					Translator:         translator,
					DetectFromTelegram: false, // Disable auto-detect, use manual language selection only
				}),
			},
		},
	)
	if err != nil {
		log.Fatalln("failed to start client:", err)
	}

	// Set the PeerStorage on the translator so language changes persist
	translator.SetSession(client.PeerStorage)

	dp := client.Dispatcher

	// Add handlers
	dp.AddHandler(handlers.OnCommand("start", start, filters.Private))
	dp.AddHandler(handlers.OnCommand("help", helpCmd, filters.Private))
	dp.AddHandler(handlers.OnCommand("language", languageCmd, filters.Private))
	dp.AddHandler(handlers.OnCallbackQuery(filters.CallbackQuery.Prefix("lang_"), languageCallback))

	fmt.Printf("Bot (@%s) has been started with i18n support...\n", client.Self.Username)

	client.Idle()
}

func start(update *adapter.Update) error {
	user := update.EffectiveUser()
	text := update.T("start", user.FirstName, "gotg bot")

	// Create language selection keyboard using gotg.Keyboard()
	kbd := gotg.Keyboard().
		Button("🇬🇧 English", "lang_en").
		Button("🇪🇸 Español", "lang_es").
		Next().
		Button(update.T("btn_menu"), "menu").
		Build()
	_, err := update.Reply(text, &adapter.SendOpts{
		ParseMode:   adapter.Markdown,
		ReplyMarkup: kbd,
	})
	return err
}

func helpCmd(update *adapter.Update) error {
	text := update.T("help")

	// Add feature list
	text += "\n\n*" + update.T("features.formatting") + "*\n"
	text += "*" + update.T("features.i18n") + "*\n"
	text += "*" + update.T("features.sessions") + "*\n"
	text += "*" + update.T("features.middleware") + "*"

	_, err := update.Reply(text, &adapter.SendOpts{
		ParseMode: adapter.Markdown,
	})
	return err
}

func languageCmd(update *adapter.Update) error {
	text := update.T("language_select")

	// Create language selection keyboard using gotg.Keyboard()
	kbd := gotg.Keyboard().
		Button("🇬🇧 English", "lang_en").
		Button("🇪🇸 Español", "lang_es").
		Build()

	_, err := update.Reply(text, &adapter.SendOpts{
		ReplyMarkup: kbd,
	})
	return err
}

func languageCallback(u *adapter.Update) error {
	var lang language.Tag
	switch u.Data() {
	case "lang_en":
		lang = language.English
	case "lang_es":
		lang = language.Spanish
	default:
		lang = language.English
	}

	// Set user's language preference using update.SetLang()
	u.SetLang(lang)

	// Get user info for greeting
	user := u.EffectiveUser()

	// Show pluralization examples
	text := u.T("language_changed", lang.String()) + "\n\n"

	// Example with count = 1
	text += u.T("items_count", &i18n.Args{
		Count: 1,
	}) + "\n"

	// Example with count = 5
	text += u.T("items_count", &i18n.Args{
		Count: 5,
	}) + "\n\n"

	// User info
	text += u.T("user_info", user.FirstName, user.ID, user.Username)

	// Answer callback query
	u.Answer(u.T("success"))

	// Try without markdown parsing first
	_, err := u.Edit(text, &adapter.EditOpts{
		ParseMode: adapter.ModeNone,
	})

	return err
}
