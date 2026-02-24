package gotg

import (
	"encoding/json"
	"fmt"

	tdSession "github.com/gotd/td/session"
	"github.com/pageton/gotg/session"
)

// loadSessionData loads the current session data from storage and unmarshals it.
func (c *Client) loadSessionData() (*tdSession.Data, error) {
	raw, err := c.sessionStorage.LoadSession(c.ctx)
	if err != nil {
		return nil, fmt.Errorf("load session: %w", err)
	}
	if len(raw) == 0 {
		return nil, fmt.Errorf("no session data available")
	}

	var v struct {
		Version int
		Data    tdSession.Data
	}
	if err := json.Unmarshal(raw, &v); err != nil {
		return nil, fmt.Errorf("unmarshal session: %w", err)
	}

	return &v.Data, nil
}

// ExportSessionAsPyrogram exports the current session as a Pyrogram-compatible string.
//
// The Pyrogram session string contains DC ID, App ID, test mode flag,
// auth key, user ID, and bot status.
//
// Note: You must not share this string with anyone, it contains auth details
// for your logged-in account.
func (c *Client) ExportSessionAsPyrogram() (string, error) {
	data, err := c.loadSessionData()
	if err != nil {
		return "", fmt.Errorf("export pyrogram: %w", err)
	}

	var userID int64
	var isBot bool
	if c.Self != nil {
		userID = c.Self.ID
		isBot = c.Self.Bot
	}

	return session.EncodePyrogramSession(data, int32(c.apiID), userID, isBot) //nolint:gosec // G115: apiID fits int32
}

// ExportSessionAsTelethon exports the current session as a Telethon-compatible string.
//
// The Telethon session string contains DC ID, IP address, port, and auth key.
//
// Note: You must not share this string with anyone, it contains auth details
// for your logged-in account.
func (c *Client) ExportSessionAsTelethon() (string, error) {
	data, err := c.loadSessionData()
	if err != nil {
		return "", fmt.Errorf("export telethon: %w", err)
	}

	return session.EncodeTelethonSession(data)
}

// ExportSessionAsGramjs exports the current session as a GramJS-compatible string.
//
// The GramJS session string contains DC ID, server address, port, and auth key.
//
// Note: You must not share this string with anyone, it contains auth details
// for your logged-in account.
func (c *Client) ExportSessionAsGramjs() (string, error) {
	data, err := c.loadSessionData()
	if err != nil {
		return "", fmt.Errorf("export gramjs: %w", err)
	}

	return session.EncodeGramjsSession(data)
}

// ExportSessionAsMtcute exports the current session as an mtcute v3-compatible string.
//
// The mtcute session string contains DC info, optional user info, and auth key
// using TL-style binary serialization with URL-safe base64 encoding.
//
// Note: You must not share this string with anyone, it contains auth details
// for your logged-in account.
func (c *Client) ExportSessionAsMtcute() (string, error) {
	data, err := c.loadSessionData()
	if err != nil {
		return "", fmt.Errorf("export mtcute: %w", err)
	}

	var userID int64
	var isBot bool
	if c.Self != nil {
		userID = c.Self.ID
		isBot = c.Self.Bot
	}

	return session.EncodeMtcuteSession(data, userID, isBot)
}
