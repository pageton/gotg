package session

import (
	"context"
	"os"

	"github.com/bytedance/sonic"
	"github.com/gotd/td/session"
	"github.com/gotd/td/session/tdesktop"
	"github.com/pageton/gotg/functions"
	"github.com/pageton/gotg/storage"
	"gorm.io/gorm"
)

type sessionName interface {
	getType() string
}

type sessionNameString string

func (sessionNameString) getType() string { return "str" }

type sessionNameDialector struct {
	dialector gorm.Dialector
}

func (sessionNameDialector) getType() string { return "dialector" }

type sessionNameAdapter struct {
	adapter storage.Adapter
}

func (sessionNameAdapter) getType() string { return "adapter" }

type SessionConstructor interface {
	loadSession() (sessionName, []byte, error)
}

type SimpleSessionConstructor int8

func SimpleSession() *SimpleSessionConstructor {
	s := SimpleSessionConstructor(0)
	return &s
}

func (*SimpleSessionConstructor) loadSession() (sessionName, []byte, error) {
	return sessionNameString("gotg_simple"), nil, nil
}

type AdapterSessionConstructor struct {
	adapter storage.Adapter
}

func WithAdapter(adapter storage.Adapter) *AdapterSessionConstructor {
	return &AdapterSessionConstructor{adapter: adapter}
}

func (s *AdapterSessionConstructor) loadSession() (sessionName, []byte, error) {
	return sessionNameAdapter{adapter: s.adapter}, nil, nil
}

type SqlSessionConstructor struct {
	dialector gorm.Dialector
}

// SqlSession creates a constructor for SQLite-based session storage.
//
// This allows loading and saving sessions from a SQLite database file.
//
// Parameters:
//   - dialector: The GORM dialector for database connection
//
// Returns:
//   - A constructor that implements SessionConstructor interface
//
// Example:
//
//	dialector := sqlite.Open("telegram.db")
//	constructor := session.SqlSession(dialector)
func SqlSession(dialector gorm.Dialector) *SqlSessionConstructor {
	return &SqlSessionConstructor{dialector: dialector}
}

func (s *SqlSessionConstructor) loadSession() (sessionName, []byte, error) {
	return &sessionNameDialector{s.dialector}, nil, nil
}

type PyrogramSessionConstructor struct {
	name, value string
}

// PyrogramSession creates a constructor for Pyrogram string session format.
//
// Pyrogram session strings use a specific hex encoding format
// ( '>BI?256sQ?' prefix followed by encoded session data).
//
// Parameters:
//   - value: The Pyrogram session string to encode
//
// Returns:
//   - A constructor that implements SessionConstructor interface
//
// Example:
//
//	constructor := session.PyrogramSession("v12345...67890abcdef")
func PyrogramSession(value string) *PyrogramSessionConstructor {
	return &PyrogramSessionConstructor{value: value}
}

func (s *PyrogramSessionConstructor) Name(name string) *PyrogramSessionConstructor {
	s.name = name
	return s
}

func (s *PyrogramSessionConstructor) loadSession() (sessionName, []byte, error) {
	sd, err := DecodePyrogramSession(s.value)
	if err != nil {
		return sessionNameString(s.name), nil, err
	}
	data, err := sonic.Marshal(jsonData{
		Version: storage.LatestVersion,
		Data:    *sd,
	})
	return sessionNameString(s.name), data, err
}

type TelethonSessionConstructor struct {
	name, value string
}

// TelethonSession creates a constructor for Telethon string session format.
//
// Telethon session strings use a specific hex encoding format
// ( '>BI?256sQ?' prefix followed by encoded session data).
//
// Parameters:
//   - value: The Telethon session string to encode
//
// Returns:
//   - A constructor that implements SessionConstructor interface
//
// Example:
//
//	constructor := session.TelethonSession("v12345...67890abcdef")
func TelethonSession(value string) *TelethonSessionConstructor {
	return &TelethonSessionConstructor{value: value}
}

func (s *TelethonSessionConstructor) Name(name string) *TelethonSessionConstructor {
	s.name = name
	return s
}

func (s *TelethonSessionConstructor) loadSession() (sessionName, []byte, error) {
	sd, err := session.TelethonSession(s.value)
	if err != nil {
		return sessionNameString(s.name), nil, err
	}
	data, err := sonic.Marshal(jsonData{
		Version: storage.LatestVersion,
		Data:    *sd,
	})
	return sessionNameString(s.name), data, err
}

type StringSessionConstructor struct {
	name, value string
}

// StringSession creates a constructor for plain string session format.
//
// This allows using simple string-encoded session data without hex encoding.
//
// Parameters:
//   - value: The session string value
//
// Returns:
//   - A constructor that implements SessionConstructor interface
//
// Example:
//
//	constructor := session.StringSession(sessionString)
func StringSession(value string) *StringSessionConstructor {
	return &StringSessionConstructor{value: value}
}

func (s *StringSessionConstructor) Name(name string) *StringSessionConstructor {
	s.name = name
	return s
}

func (s *StringSessionConstructor) loadSession() (sessionName, []byte, error) {
	sd, err := functions.DecodeStringToSession(s.value)
	if err != nil {
		return sessionNameString(s.name), nil, err
	}
	return sessionNameString(s.name), sd.Data, err
}

type TdataSessionConstructor struct {
	Account tdesktop.Account
	name    string
}

// TdataSession creates a constructor for Telegram Desktop session format.
//
// Telegram Desktop session uses a specific binary format for session storage.
//
// Parameters:
//   - account: The tdesktop.Account containing session data
//
// Returns:
//   - A constructor that implements SessionConstructor interface
//
// Example:
//
//	constructor := session.TdataSession(account)
func TdataSession(account tdesktop.Account) *TdataSessionConstructor {
	return &TdataSessionConstructor{Account: account}
}

func (s *TdataSessionConstructor) Name(name string) *TdataSessionConstructor {
	s.name = name
	return s
}

func (s *TdataSessionConstructor) loadSession() (sessionName, []byte, error) {
	sd, err := session.TDesktopSession(s.Account)
	if err != nil {
		return sessionNameString(s.name), nil, err
	}
	ctx := context.Background()
	var (
		gotdstorage = new(session.StorageMemory)
		loader      = session.Loader{Storage: gotdstorage}
	)
	// Save decoded Telegram Desktop session as gotd session.
	if err := loader.Save(ctx, sd); err != nil {
		return sessionNameString(s.name), nil, err
	}
	data, err := sonic.Marshal(jsonData{
		Version: storage.LatestVersion,
		Data:    *sd,
	})
	return sessionNameString(s.name), data, err
}

type GramjsSessionConstructor struct {
	name, value string
}

// GramjsSession creates a constructor for Gram.js string session format.
//
// Gram.js session strings use a specific hex encoding format.
//
// Parameters:
//   - value: The Gram.js session string to encode
//
// Returns:
//   - A constructor that implements SessionConstructor interface
//
// Example:
//
//	constructor := session.GramjsSession("v12345...67890abcdef")
func GramjsSession(value string) *GramjsSessionConstructor {
	return &GramjsSessionConstructor{value: value}
}

func (s *GramjsSessionConstructor) Name(name string) *GramjsSessionConstructor {
	s.name = name
	return s
}

func (s *GramjsSessionConstructor) loadSession() (sessionName, []byte, error) {
	sd, err := DecodeGramjsSession(s.value)
	if err != nil {
		return sessionNameString(s.name), nil, err
	}
	data, err := sonic.Marshal(jsonData{
		Version: storage.LatestVersion,
		Data:    *sd,
	})
	return sessionNameString(s.name), data, err
}

type MtcuteSessionConstructor struct {
	name, value string
}

// MtcuteSession creates a constructor for mtcute string session format.
//
// mtcute session strings use URL-safe base64 encoding with TL-style
// binary serialization (version 3).
//
// Parameters:
//   - value: The mtcute session string to decode
//
// Returns:
//   - A constructor that implements SessionConstructor interface
//
// Example:
//
//	constructor := session.MtcuteSession("AwQAAAAXAgIA...")
func MtcuteSession(value string) *MtcuteSessionConstructor {
	return &MtcuteSessionConstructor{value: value}
}

func (s *MtcuteSessionConstructor) Name(name string) *MtcuteSessionConstructor {
	s.name = name
	return s
}

func (s *MtcuteSessionConstructor) loadSession() (sessionName, []byte, error) {
	sd, err := DecodeMtcuteSession(s.value)
	if err != nil {
		return sessionNameString(s.name), nil, err
	}
	data, err := sonic.Marshal(jsonData{
		Version: storage.LatestVersion,
		Data:    *sd,
	})
	return sessionNameString(s.name), data, err
}

type JsonFileSessionConstructor struct {
	name, filePath string
}

// JsonFileSession creates a constructor for JSON file session format.
//
// This allows loading and saving sessions from a local JSON file.
//
// Parameters:
//   - filePath: Path to the JSON session file
//
// Returns:
//   - A constructor that implements SessionConstructor interface
//
// Example:
//
//	constructor := session.JsonFileSession("/path/to/session.json")
func JsonFileSession(filePath string) *JsonFileSessionConstructor {
	return &JsonFileSessionConstructor{filePath: filePath}
}

func (s *JsonFileSessionConstructor) Name(name string) *JsonFileSessionConstructor {
	s.name = name
	return s
}

func (s *JsonFileSessionConstructor) loadSession() (sessionName, []byte, error) {
	buf, err := os.ReadFile(s.filePath)
	return sessionNameString(s.name), buf, err
}
