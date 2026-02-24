package adapter

import (
	"github.com/gotd/td/tg"
	"github.com/pageton/gotg/functions"
)

// Enable2FA enables Two-Factor Authentication by setting a new cloud password.
//
// See https://core.telegram.org/api/srp#setting-a-new-2fa-password for reference.
func (ctx *Context) Enable2FA(newPassword string, opts *functions.PasswordOpts) error {
	return functions.Enable2FA(ctx.Context, ctx.Raw, newPassword, opts)
}

// Update2FA changes an existing 2FA cloud password to a new one.
//
// See https://core.telegram.org/api/srp#setting-a-new-2fa-password for reference.
func (ctx *Context) Update2FA(currentPassword, newPassword string, opts *functions.PasswordOpts) error {
	return functions.Update2FA(ctx.Context, ctx.Raw, currentPassword, newPassword, opts)
}

// Disable2FA removes the 2FA cloud password from the account.
//
// See https://core.telegram.org/api/srp for reference.
func (ctx *Context) Disable2FA(currentPassword string) error {
	return functions.Disable2FA(ctx.Context, ctx.Raw, currentPassword)
}

// GetActiveSessions returns all active authorized sessions for the current account.
//
// See https://core.telegram.org/method/account.getAuthorizations for reference.
func (ctx *Context) GetActiveSessions() ([]tg.Authorization, error) {
	return functions.GetActiveSessions(ctx.Context, ctx.Raw)
}

// RevokeSession terminates an active authorized session by its hash.
//
// See https://core.telegram.org/method/account.resetAuthorization for reference.
func (ctx *Context) RevokeSession(hash int64) error {
	return functions.RevokeSession(ctx.Context, ctx.Raw, hash)
}

// RevokeAllOtherSessions terminates all other active sessions except the current one.
//
// See https://core.telegram.org/method/auth.resetAuthorizations for reference.
func (ctx *Context) RevokeAllOtherSessions() error {
	return functions.RevokeAllOtherSessions(ctx.Context, ctx.Raw)
}
