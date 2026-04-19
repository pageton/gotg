package storage

import (
	"database/sql/driver"

	"github.com/bytedance/sonic"
)

// Usernames is a named slice of Username, enabling database serialization.
type Usernames []Username

// MarshalJSON serializes Username as an object.
func (u Username) MarshalJSON() ([]byte, error) {
	type alias Username
	return sonic.Marshal(alias(u))
}

// UnmarshalJSON deserializes Username from either:
//   - new format: {"Username":"alice","Active":true,"Editable":true}
//   - legacy format: "alice" (plain string from old gotg versions)
func (u *Username) UnmarshalJSON(data []byte) error {
	// Try plain string first (legacy format).
	var s string
	if err := sonic.Unmarshal(data, &s); err == nil {
		u.Username = s
		u.Active = true
		u.Editable = true
		return nil
	}
	// Otherwise, decode as object.
	type alias Username
	var a alias
	if err := sonic.Unmarshal(data, &a); err != nil {
		return err
	}
	*u = Username(a)
	return nil
}

// Value implements driver.Valuer for database storage (JSON serialization).
func (u Usernames) Value() (driver.Value, error) {
	if u == nil {
		return nil, nil
	}
	return sonic.Marshal(u)
}

// Scan implements sql.Scanner for database retrieval (JSON deserialization).
func (u *Usernames) Scan(val interface{}) error {
	if val == nil {
		*u = nil
		return nil
	}
	var bytes []byte
	switch v := val.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return nil
	}
	return sonic.Unmarshal(bytes, u)
}
