package adapter

import (
	"fmt"
	"html"

	"github.com/gotd/td/tg"
	"github.com/pageton/gotg/functions"
)

// Mention generates an HTML mention link for a Telegram user.
//
// Behavior:
// - No arguments: uses the Update's default UserID() and FullName().
// - One argument:
//   - int/int64 → overrides userID, keeps default name.
//   - string → overrides name, keeps default userID.
//
// - Two arguments: first is userID (int/int64), second is name (string).
// - The name can be any string, including numeric names.
// - Returns a string in the format: <a href='tg://user?id=USERID'>NAME</a>
func (u *Update) Mention(args ...any) string {
	userID := u.UserID()
	name := u.FullName()

	if len(args) == 1 {
		switch v := args[0].(type) {
		case int:
			userID = int64(v)
		case int64:
			userID = v
		case string:
			name = html.EscapeString(v)
		}
	} else if len(args) >= 2 {
		switch v := args[0].(type) {
		case int:
			userID = int64(v)
		case int64:
			userID = v
		}
		if n, ok := args[1].(string); ok {
			name = html.EscapeString(n)
		}
	}

	return fmt.Sprintf("<a href='tg://user?id=%d'>%s</a>", userID, name)
}

// Delete deletes the effective message for this update.
// Returns an error if the deletion fails.
func (u *Update) Delete() error {
	return u.Ctx.DeleteMessages(u.ChatID(), []int{u.MsgID()})
}

// GetFullUser fetches full user information for the effective user.
// Returns nil if no user exists or on error.
func (u *Update) GetFullUser() (*tg.UserFull, error) {
	return u.Ctx.GetFullUser(u.UserID())
}

// GetUser fetches user information for the effective user.
// Returns nil if no user exists or on error.
func (u *Update) GetUser() (*tg.User, error) {
	return u.Ctx.GetUser(u.UserID())
}

// Pin pins the effective message in the chat.
// Returns updates confirming the action or an error.
func (u *Update) Pin() (tg.UpdatesClass, error) {
	return u.Ctx.PinMessage(u.ChatID(), u.MsgID())
}

// UnPin unpins the effective message in the chat.
// Returns an error if the operation fails.
func (u *Update) UnPin() error {
	return u.Ctx.UnPinMessage(u.ChatID(), u.MsgID())
}

// UnPinAll unpins all messages in the current chat.
// Returns an error if the operation fails.
func (u *Update) UnPinAll() error {
	return u.Ctx.UnPinAllMessages(u.ChatID())
}

// T returns a translation for the given key.
// Supports both simple args and context (Args) for pluralization, gender, etc.
// This method requires i18n middleware to be initialized.
//
// Examples:
//
//	// Simple translation with positional args
//	text := u.T("greeting", userName)
//
//	// Translation with context (pluralization, gender)
//	text := u.T("items_count", &i18n.Args{Count: 5})
//
//	// Translation with named args
//	text := u.T("welcome", &i18n.Args{Args: map[string]any{"name": userName}})
func (u *Update) T(key string, args ...any) string {
	if u.Ctx == nil {
		return key
	}
	return updateTImpl(u, key, args...)
}

// SetLang sets the language preference for the effective user.
// This requires i18n middleware to be initialized.
//
// Parameters:
//   - lang: The language code (e.g., "en", "es") or language.Tag
func (u *Update) SetLang(lang any) {
	updateSetLangImpl(u, lang)
}

// GetLang returns the user's current language preference.
// This method requires i18n middleware to be initialized.
//
// Example:
//
//	lang := u.GetLang()
func (u *Update) GetLang() any {
	return updateGetLangImpl(u)
}

// GetChatInviteLink generates an invite link for a chat.
//
// Parameters:
//   - chatID: The chat ID to generate invite link for (use 0 for current update's chat)
//   - req: Telegram's MessagesExportChatInviteRequest (optional, use &tg.MessagesExportChatInviteRequest{} for default)
//
// Returns exported chat invite or an error.
func (u *Update) GetChatInviteLink(req ...*tg.MessagesExportChatInviteRequest) (tg.ExportedChatInviteClass, error) {
	chatID := u.ChatID()
	if chatID == 0 {
		return nil, fmt.Errorf("no chat found")
	}
	return u.Ctx.GetChatInviteLink(chatID, req...)
}

// DumpValue returns the raw UpdateClass for JSON serialization.
func (u *Update) DumpValue() any {
	return u.UpdateClass
}

// Dump returns pretty-printed JSON of any value with optional key prefix.
func (u *Update) Dump(val any, key ...string) string {
	return functions.Dump(val, key...)
}
