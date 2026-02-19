package gotg

import (
	"context"
	"fmt"
	"runtime"
	"sync"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/telegram/updates"
	"github.com/gotd/td/telegram/updates/hook"
	"github.com/gotd/td/tg"
	"github.com/pageton/gotg/adapter"
	intErrors "github.com/pageton/gotg/errors"
	"github.com/pageton/gotg/functions"
	"github.com/pageton/gotg/storage"
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
	c.deviceParams = device.Params

	// Create the gap manager that sits between gotd's raw update stream
	// and gotg's dispatcher. It tracks pts/qts/seq, detects gaps, and
	// automatically calls getDifference/getChannelDifference to recover
	// missed updates (messages, callback queries, etc.).
	c.gapManager = updates.New(updates.Config{
		Handler: c.Dispatcher,
		Logger:  c.Logger.ZapLogger(),
	})

	// The update hook middleware feeds API-response updates (e.g. from
	// messages.sendMessage returning UpdatesClass) into the gap manager
	// so their pts/seq values are tracked correctly.
	gapMiddleware := hook.UpdateHook(func(ctx context.Context, u tg.UpdatesClass) error {
		return c.gapManager.Handle(ctx, u)
	})
	middlewares = append(middlewares, gapMiddleware)

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
		UpdateHandler:     c.gapManager,
		NoUpdates:         c.NoUpdates,
		SessionStorage:    c.sessionStorage,
		Logger:            c.Logger.ZapLogger(),
		Device:            *device,
		Middlewares:       middlewares,
	})
}

func (c *Client) login() error {
	authClient := c.Auth()
	status, err := authClient.Status(c.ctx)
	if err != nil {
		return fmt.Errorf("auth status: %w", err)
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
		var flowClient auth.FlowClient = authClient
		if solver, ok := c.authConversator.(RecaptchaSolver); ok {
			flowClient = FlowClient{
				FlowClient: authClient,
				api:        c.API(),
				apiID:      c.apiID,
				apiHash:    c.apiHash,
				params:     c.deviceParams,
				solver:     solver,
			}
		}
		err = authFlow(
			c.ctx, flowClient,
			c.authConversator,
			c.clientType.getValue(),
			c.sendCodeOptions,
		)
		if err != nil {
			return fmt.Errorf("auth flow: %w", err)
		}
	} else {
		if !status.Authorized && c.clientType.getValue() != "" {
			if _, err := c.Auth().Bot(c.ctx, c.clientType.getValue()); err != nil {
				return fmt.Errorf("login: %w", err)
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

		if !c.NoUpdates {
			return c.gapManager.Run(ctx, c.API(), self.ID, updates.AuthOptions{
				IsBot:  self.Bot,
				Forget: false,
			})
		}

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
	ctx := adapter.NewContext(
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
		nil,
		c.Client,
	)
	ctx.DefaultParseMode = c.defaultParseMode
	return ctx
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
	// Drain pending DB writes before tearing down.
	if c.PeerStorage != nil {
		c.PeerStorage.Close()
	}

	// Wait for in-flight updates and close DC pools.
	if c.Dispatcher != nil {
		c.Dispatcher.WaitPending()
		c.Dispatcher.CloseDCPools()
	}

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
		defer func() {
			if r := recover(); r != nil {
				c.err = fmt.Errorf("panic in client: %v", r)
				wg.Done()
			}
		}()
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
