package gotg

import (
	"context"
	"fmt"
	"runtime"
	"sync"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/tg"
	"github.com/pageton/gotg/adapter"
	intErrors "github.com/pageton/gotg/errors"
	"github.com/pageton/gotg/functions"
	"github.com/pageton/gotg/storage"
	"github.com/pkg/errors"
)

func (c *Client) initTelegramClient(
	device *telegram.DeviceConfig,
	middlewares []telegram.Middleware,
) {
	if device == nil {
		device = &telegram.DeviceConfig{
			DeviceModel:    "gotg",
			SystemVersion:  runtime.GOOS,
			AppVersion:     VERSION,
			SystemLangCode: c.SystemLangCode,
			LangCode:       c.ClientLangCode,
		}
	}
	c.Client = telegram.NewClient(c.apiID, c.apiHash, telegram.Options{
		DCList:            c.DCList,
		Resolver:          c.Resolver,
		DC:                c.DC,
		PublicKeys:        c.PublicKeys,
		MigrationTimeout:  c.MigrationTimeout,
		AckBatchSize:      c.AckBatchSize,
		AckInterval:       c.AckInterval,
		RetryInterval:     c.RetryInterval,
		MaxRetries:        c.MaxRetries,
		ExchangeTimeout:   c.ExchangeTimeout,
		DialTimeout:       c.DialTimeout,
		CompressThreshold: c.CompressThreshold,
		UpdateHandler:     c.Dispatcher,
		NoUpdates:         c.NoUpdates,
		SessionStorage:    c.sessionStorage,
		Logger:            c.Logger,
		Device:            *device,
		Middlewares:       middlewares,
	})
}

func (c *Client) login() error {
	authClient := c.Auth()
	status, err := authClient.Status(c.ctx)
	if err != nil {
		return errors.Wrap(err, "auth status")
	}
	if status.Authorized {
		return nil
	}
	if c.clientType.getType() == clientTypeVPhone {
		if c.NoAutoAuth {
			return intErrors.ErrSessionUnauthorized
		}
		if c.clientType.getValue() == "" {
			return intErrors.ErrSessionUnauthorized
		}
		err = authFlow(
			c.ctx, authClient,
			c.authConversator,
			c.clientType.getValue(),
			c.sendCodeOptions,
		)
		if err != nil {
			return errors.Wrap(err, "auth flow")
		}
	} else {
		if !status.Authorized && c.clientType.getValue() != "" {
			if _, err := c.Auth().Bot(c.ctx, c.clientType.getValue()); err != nil {
				return errors.Wrap(err, "login")
			}
		}
	}
	return nil
}

func (ch *Client) printCredit() {
	if !ch.DisableCopyright {
		fmt.Printf(`
gotg %s, Copyright (C) 2026 Sadiq <github.com/pageton>
Licensed under the terms of GNU General Public License v3

`, VERSION)
	}
}

func (c *Client) initialize(wg *sync.WaitGroup) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		err := c.login()
		if err != nil {
			return err
		}
		self, err := c.Client.Self(ctx)
		if err != nil {
			return err
		}

		c.Self = self

		c.Dispatcher.Initialize(ctx, c.Stop, c.Client, self)

		c.PeerStorage.AddPeer(self.ID, self.AccessHash, storage.TypeUser, self.Username)
		wg.Done()
		c.running = true
		<-c.ctx.Done()
		return c.ctx.Err()
	}
}

// ExportStringSession EncodeSessionToString encodes the client session to a string in base64.
//
// Note: You must not share this string with anyone, it contains auth details for your logged in account.
func (c *Client) ExportStringSession() (string, error) {
	loadedSessionData, err := c.sessionStorage.LoadSession(c.ctx)
	if err == nil {
		loadedSession := &storage.Session{
			Version: storage.LatestVersion,
			Data:    loadedSessionData,
		}
		return functions.EncodeSessionToString(loadedSession)
	}
	return functions.EncodeSessionToString(c.PeerStorage.GetSession())
}

// Idle keeps the current goroutined blocked until the client is stopped.
func (c *Client) Idle() error {
	<-c.ctx.Done()
	return c.err
}

// CreateContext creates a new pseudo updates context.
// A context retrieved from this method should be reused.
func (c *Client) CreateContext() *adapter.Context {
	return adapter.NewContext(
		c.ctx,
		c.API(),
		c.PeerStorage,
		c.Self,
		message.NewSender(c.API()),
		&tg.Entities{
			Users: map[int64]*tg.User{
				c.Self.ID: c.Self,
			},
		},
		c.autoFetchReply,
		c.ConvManager,
	)
}

// Stop cancels the context.Context being used for the client
// and stops it.
//
// Notes:
//
// 1.) Client.Idle() will exit if this method is called.
//
// 2.) You can call Client.Start() to start the client again
// if it was stopped using this method.
func (c *Client) Stop() {
	c.cancel()
	c.running = false
}

// Start connects the client to telegram servers and logins.
// It will return error if the client is already running.
func (c *Client) Start(opts *ClientOpts) error {
	if c.running {
		return intErrors.ErrClientAlreadyRunning
	}
	if c.ctx.Err() == context.Canceled {
		c.ctx, c.cancel = context.WithCancel(context.Background())
	}

	c.initTelegramClient(opts.Device, opts.Middlewares)
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func(c *Client) {
		if opts.RunMiddleware == nil {
			c.err = c.Run(c.ctx, c.initialize(&wg))
		} else {
			c.err = opts.RunMiddleware(
				c.Run,
				c.ctx,
				c.initialize(&wg),
			)
		}

		if c.err != nil {
			wg.Done()
		}
	}(c)

	wg.Wait()
	if c.err == nil {
		if !c.Self.Bot && opts.PeersFromDialogs {
			if opts.WaitOnPeersFromDialogs {
				storage.AddPeersFromDialogs(c.ctx, c.API(), c.PeerStorage)
			} else {
				go storage.AddPeersFromDialogs(c.ctx, c.API(), c.PeerStorage)
			}
		}
	}
	return c.err
}

// RefreshContext casts the new context.Context and telegram session
// to ext.Context (It may be used after doing Stop and Start calls respectively.)
func (c *Client) RefreshContext(ctx *adapter.Context) {
	(*ctx).Context = c.ctx
	(*ctx).Raw = c.API()
}
