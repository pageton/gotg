package functions

import (
	"context"
	"fmt"

	"github.com/gotd/td/crypto"
	"github.com/gotd/td/crypto/srp"
	"github.com/gotd/td/tg"
)

// PasswordOpts holds optional parameters for 2FA password operations.
type PasswordOpts struct {
	// Hint is a text hint for the password, shown when the user
	// is asked to enter it during login.
	Hint string
	// Email is the recovery email address associated with the 2FA password.
	// Telegram will send a confirmation code to this email.
	Email string
}

// Enable2FA enables Two-Factor Authentication by setting a new cloud password.
//
// This should only be called when the account does not already have 2FA enabled.
// Use Update2FA to change an existing password.
//
// Parameters:
//   - ctx: Context for the API call
//   - raw: The raw Telegram client
//   - newPassword: The new 2FA password to set
//   - opts: Optional hint and recovery email (can be nil)
//
// Returns an error if 2FA is already enabled or the operation fails.
//
// See https://core.telegram.org/api/srp#setting-a-new-2fa-password for reference.
func Enable2FA(ctx context.Context, raw *tg.Client, newPassword string, opts *PasswordOpts) error {
	p, err := raw.AccountGetPassword(ctx)
	if err != nil {
		return fmt.Errorf("get password parameters: %w", err)
	}
	if p.HasPassword {
		return fmt.Errorf("2FA is already enabled, use Update2FA to change the password")
	}

	algo, ok := p.NewAlgo.(*tg.PasswordKdfAlgoSHA256SHA256PBKDF2HMACSHA512iter100000SHA256ModPow)
	if !ok {
		return fmt.Errorf("unsupported password algo: %T", p.NewAlgo)
	}

	newHash, err := computeNewPasswordHash([]byte(newPassword), algo)
	if err != nil {
		return fmt.Errorf("compute new password hash: %w", err)
	}

	settings := tg.AccountPasswordInputSettings{}
	settings.SetNewAlgo(algo)
	settings.SetNewPasswordHash(newHash)
	if opts != nil {
		if opts.Hint != "" {
			settings.SetHint(opts.Hint)
		}
		if opts.Email != "" {
			settings.SetEmail(opts.Email)
		}
	}

	if _, err := raw.AccountUpdatePasswordSettings(ctx, &tg.AccountUpdatePasswordSettingsRequest{
		Password:    &tg.InputCheckPasswordEmpty{},
		NewSettings: settings,
	}); err != nil {
		return fmt.Errorf("enable 2FA: %w", err)
	}
	return nil
}

// Update2FA changes an existing 2FA cloud password to a new one.
//
// The current password is required for verification via SRP.
//
// Parameters:
//   - ctx: Context for the API call
//   - raw: The raw Telegram client
//   - currentPassword: The current 2FA password for verification
//   - newPassword: The new 2FA password to set
//   - opts: Optional hint and recovery email (can be nil)
//
// Returns an error if 2FA is not enabled or the operation fails.
//
// See https://core.telegram.org/api/srp#setting-a-new-2fa-password for reference.
func Update2FA(ctx context.Context, raw *tg.Client, currentPassword, newPassword string, opts *PasswordOpts) error {
	p, err := raw.AccountGetPassword(ctx)
	if err != nil {
		return fmt.Errorf("get password parameters: %w", err)
	}
	if !p.HasPassword {
		return fmt.Errorf("2FA is not enabled, use Enable2FA first")
	}

	algo, ok := p.NewAlgo.(*tg.PasswordKdfAlgoSHA256SHA256PBKDF2HMACSHA512iter100000SHA256ModPow)
	if !ok {
		return fmt.Errorf("unsupported password algo: %T", p.NewAlgo)
	}

	newHash, err := computeNewPasswordHash([]byte(newPassword), algo)
	if err != nil {
		return fmt.Errorf("compute new password hash: %w", err)
	}

	oldSRP, err := computePasswordSRP([]byte(currentPassword), p)
	if err != nil {
		return fmt.Errorf("compute current password SRP: %w", err)
	}

	settings := tg.AccountPasswordInputSettings{}
	settings.SetNewAlgo(algo)
	settings.SetNewPasswordHash(newHash)
	if opts != nil {
		if opts.Hint != "" {
			settings.SetHint(opts.Hint)
		}
		if opts.Email != "" {
			settings.SetEmail(opts.Email)
		}
	}

	if _, err := raw.AccountUpdatePasswordSettings(ctx, &tg.AccountUpdatePasswordSettingsRequest{
		Password:    oldSRP,
		NewSettings: settings,
	}); err != nil {
		return fmt.Errorf("update 2FA: %w", err)
	}
	return nil
}

// Disable2FA removes the 2FA cloud password from the account.
//
// The current password is required for verification via SRP.
//
// Parameters:
//   - ctx: Context for the API call
//   - raw: The raw Telegram client
//   - currentPassword: The current 2FA password for verification
//
// Returns an error if 2FA is not enabled or the operation fails.
//
// See https://core.telegram.org/api/srp for reference.
func Disable2FA(ctx context.Context, raw *tg.Client, currentPassword string) error {
	p, err := raw.AccountGetPassword(ctx)
	if err != nil {
		return fmt.Errorf("get password parameters: %w", err)
	}
	if !p.HasPassword {
		return fmt.Errorf("2FA is not enabled")
	}

	oldSRP, err := computePasswordSRP([]byte(currentPassword), p)
	if err != nil {
		return fmt.Errorf("compute current password SRP: %w", err)
	}

	settings := tg.AccountPasswordInputSettings{}
	settings.SetNewAlgo(&tg.PasswordKdfAlgoUnknown{})
	settings.SetNewPasswordHash([]byte{})
	settings.SetHint("")

	if _, err := raw.AccountUpdatePasswordSettings(ctx, &tg.AccountUpdatePasswordSettingsRequest{
		Password:    oldSRP,
		NewSettings: settings,
	}); err != nil {
		return fmt.Errorf("disable 2FA: %w", err)
	}
	return nil
}

// computePasswordSRP computes the SRP answer for an existing password.
func computePasswordSRP(password []byte, p *tg.AccountPassword) (*tg.InputCheckPasswordSRP, error) {
	s := srp.NewSRP(crypto.DefaultRand())

	algo, ok := p.CurrentAlgo.(*tg.PasswordKdfAlgoSHA256SHA256PBKDF2HMACSHA512iter100000SHA256ModPow)
	if !ok {
		return nil, fmt.Errorf("unsupported current algo: %T", p.CurrentAlgo)
	}

	a, err := s.Hash(password, p.SRPB, p.SecureRandom, srp.Input(*algo))
	if err != nil {
		return nil, fmt.Errorf("compute SRP hash: %w", err)
	}

	return &tg.InputCheckPasswordSRP{
		SRPID: p.SRPID,
		A:     a.A,
		M1:    a.M1,
	}, nil
}

// computeNewPasswordHash computes the hash for a new password being set.
// Mutates algo.Salt1 as required by the SRP protocol.
func computeNewPasswordHash(
	password []byte,
	algo *tg.PasswordKdfAlgoSHA256SHA256PBKDF2HMACSHA512iter100000SHA256ModPow,
) ([]byte, error) {
	s := srp.NewSRP(crypto.DefaultRand())

	hash, newSalt, err := s.NewHash(password, srp.Input(*algo))
	if err != nil {
		return nil, fmt.Errorf("compute new hash: %w", err)
	}
	algo.Salt1 = newSalt

	return hash, nil
}

// GetActiveSessions returns all active authorized sessions for the current account.
//
// See https://core.telegram.org/method/account.getAuthorizations for reference.
func GetActiveSessions(ctx context.Context, raw *tg.Client) ([]tg.Authorization, error) {
	result, err := raw.AccountGetAuthorizations(ctx)
	if err != nil {
		return nil, fmt.Errorf("get active sessions: %w", err)
	}
	return result.Authorizations, nil
}

// RevokeSession terminates an active authorized session by its hash.
//
// The session hash can be obtained from GetActiveSessions.
// Cannot revoke the current session — use auth.logOut for that.
//
// See https://core.telegram.org/method/account.resetAuthorization for reference.
func RevokeSession(ctx context.Context, raw *tg.Client, hash int64) error {
	if _, err := raw.AccountResetAuthorization(ctx, hash); err != nil {
		return fmt.Errorf("revoke session: %w", err)
	}
	return nil
}

// RevokeAllOtherSessions terminates all other active sessions except the current one.
//
// See https://core.telegram.org/method/auth.resetAuthorizations for reference.
func RevokeAllOtherSessions(ctx context.Context, raw *tg.Client) error {
	if _, err := raw.AuthResetAuthorizations(ctx); err != nil {
		return fmt.Errorf("revoke all other sessions: %w", err)
	}
	return nil
}
