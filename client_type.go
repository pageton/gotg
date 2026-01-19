package gotg

const (
	clientTypeVPhone int = iota
	clientTypeVBot
)

type clientType interface {
	getType() int
	getValue() string
}

type clientTypePhone string

func (v *clientTypePhone) getType() int {
	return clientTypeVPhone
}

func (v clientTypePhone) getValue() string {
	return string(v)
}

func ClientTypePhone(phoneNumber string) clientType {
	v := clientTypePhone(phoneNumber)
	return &v
}

// AsUser is a simpler alias for ClientTypePhone
func AsUser(phoneNumber string) clientType {
	return ClientTypePhone(phoneNumber)
}

type clientTypeBot string

func (v *clientTypeBot) getType() int {
	return clientTypeVBot
}

func (v clientTypeBot) getValue() string {
	return string(v)
}

func ClientTypeBot(botToken string) clientType {
	v := clientTypeBot(botToken)
	return &v
}

// AsBot is a simpler alias for ClientTypeBot
func AsBot(botToken string) clientType {
	return ClientTypeBot(botToken)
}

type clientTypeSimple struct{}

func (v *clientTypeSimple) getType() int {
	return clientTypeVPhone // Treat as user type
}

func (v clientTypeSimple) getValue() string {
	return ""
}

// Simple creates a client type for use with string sessions.
// No phone number or bot token is needed when using this type.
func Simple() clientType {
	return &clientTypeSimple{}
}
