package gotg

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgerr"
	gotgErrors "github.com/pageton/gotg/errors"
)

// Flow wraps auth.Flow to add custom sign-up handling.
//
// It extends the standard Telegram authentication flow with support for
// interactive sign-up prompts when required by Telegram's API.
// Flow wraps auth.Flow to add custom sign-up handling.
// It extends the standard Telegram authentication flow with support for
// interactive sign-up prompts when required.
type Flow auth.Flow

func (f Flow) handleSignUp(ctx context.Context, client auth.FlowClient, phone, hash string, s *auth.SignUpRequired) error {
	if err := f.Auth.AcceptTermsOfService(ctx, s.TermsOfService); err != nil {
		return fmt.Errorf("confirm TOS: %w", err)
	}
	info, err := f.Auth.SignUp(ctx)
	if err != nil {
		return fmt.Errorf("sign up info not provided: %w", err)
	}
	if _, err := client.SignUp(ctx, auth.SignUp{
		PhoneNumber:   phone,
		PhoneCodeHash: hash,
		FirstName:     info.FirstName,
		LastName:      info.LastName,
	}); err != nil {
		return fmt.Errorf("sign up: %w", err)
	}
	return nil
}

func authFlow(ctx context.Context, client auth.FlowClient, conversator AuthConversator, phone string, sendOpts auth.SendCodeOptions) error {
	f := Flow(auth.NewFlow(
		termAuth{
			phone:       phone,
			conversator: conversator,
		},
		sendOpts,
	))
	if f.Auth == nil {
		return gotgErrors.ErrNoAuthenticator
	}

	var (
		sentCode tg.AuthSentCodeClass
		err      error
	)
	SendAuthStatus(conversator, AuthStatusPhoneAsked)
	for i := range 3 {
		if i > 0 {
			time.Sleep(time.Duration(i) * time.Second)
		}
		var err1 error
		if i == 0 {
			phone, err1 = f.Auth.Phone(ctx)
		} else {
			SendAuthStatusWithRetrials(conversator, AuthStatusPhoneRetrial, 3-i)
			phone, err1 = conversator.AskPhoneNumber()
		}
		if err1 != nil {
			return fmt.Errorf("get phone: %w", err1)
		}
		sentCode, err = client.SendCode(ctx, phone, f.Options)
		if tgerr.Is(err, "PHONE_NUMBER_INVALID") {
			continue
		}
		break
	}
	if err != nil {
		SendAuthStatus(conversator, AuthStatusPhoneFailed)
		return err
	}

	// phone, err := f.Auth.Phone(ctx)
	// if err != nil {
	// 	return errors.Wrap(err, "get phone")
	// }

	// sentCode, err := client.SendCode(ctx, phone, f.Options)
	// if err != nil {
	// 	return err
	// }
	switch s := sentCode.(type) {
	case *tg.AuthSentCode:
		hash := s.PhoneCodeHash
		var signInErr error
		for i := range 3 {
			if i > 0 {
				time.Sleep(time.Duration(i) * time.Second)
			}
			var code string
			if i == 0 {
				conversator.AuthStatus(AuthStatus{
					Event:        AuthStatusPhoneCodeAsked,
					SentCodeType: s.Type,
				})
				code, err = f.Auth.Code(ctx, s)
			} else {
				SendAuthStatusWithRetrials(conversator, AuthStatusPhoneCodeRetrial, 3-i)
				code, err = conversator.AskCode()
			}
			if err != nil {
				SendAuthStatus(conversator, AuthStatusPhoneCodeFailed)
				return fmt.Errorf("get code: %w", err)
			}
			_, signInErr = client.SignIn(ctx, phone, code, hash)
			if tgerr.Is(signInErr, "PHONE_CODE_INVALID") {
				continue
			}
			break
		}
		// code, err := f.Auth.Code(ctx, s)
		// if err != nil {
		// 	return errors.Wrap(err, "get code")
		// }
		// _, signInErr := client.SignIn(ctx, phone, code, hash)

		if errors.Is(signInErr, auth.ErrPasswordAuthNeeded) {
			SendAuthStatus(conversator, AuthStatusPasswordAsked)
			err = signInErr
			for i := 0; err != nil && i < 3; i++ {
				if i > 0 {
					time.Sleep(time.Duration(i) * time.Second)
				}
				var password string
				var err1 error
				if i == 0 {
					password, err1 = f.Auth.Password(ctx)
				} else {
					SendAuthStatusWithRetrials(conversator, AuthStatusPasswordRetrial, 3-i)
					password, err1 = conversator.AskPassword()
				}
				if err1 != nil {
					return fmt.Errorf("get password: %w", err1)
				}
				_, err = client.Password(ctx, password)
				if err == auth.ErrPasswordInvalid {
					continue
				}
				break
			}
			if err != nil {
				SendAuthStatus(conversator, AuthStatusPasswordFailed)
				return fmt.Errorf("sign in with password: %w", err)
			}
			SendAuthStatus(conversator, AuthStatusSuccess)
			return nil
		}
		var signUpRequired *auth.SignUpRequired
		if errors.As(signInErr, &signUpRequired) {
			return f.handleSignUp(ctx, client, phone, hash, signUpRequired)
		}
		if signInErr != nil {
			SendAuthStatus(conversator, AuthStatusPhoneCodeFailed)
			return fmt.Errorf("sign in: %w", signInErr)
		}
		SendAuthStatus(conversator, AuthStatusSuccess)
	case *tg.AuthSentCodeSuccess:
		switch a := s.Authorization.(type) {
		case *tg.AuthAuthorization:
			SendAuthStatus(conversator, AuthStatusSuccess)
			// Looks that we are already authorized.
			return nil
		case *tg.AuthAuthorizationSignUpRequired:
			if err := f.handleSignUp(ctx, client, phone, "", &auth.SignUpRequired{
				TermsOfService: a.TermsOfService,
			}); err != nil {
				// TODO: not sure that blank hash will work here
				return fmt.Errorf("sign up after auth sent code success: %w", err)
			}
			return nil
		default:
			return fmt.Errorf("unexpected authorization type: %T", a)
		}
	default:
		return fmt.Errorf("unexpected sent code type: %T", sentCode)
	}

	return nil
}
