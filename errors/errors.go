package errors
import "errors"

var (
	ErrClientAlreadyRunning  = errors.New("client is already running")
	ErrSessionUnauthorized   = errors.New("session is unauthorized")
	ErrConversationCancelled = errors.New("conversation cancelled")
	ErrConversationTimeout   = errors.New("conversation: timeout waiting for response")
	ErrConversationClosed    = errors.New("conversation: closed")
)

var (
	ErrPeerNotFound       = errors.New("peer not found")
	ErrNotChat            = errors.New("not chat")
	ErrNotChannel         = errors.New("not channel")
	ErrNotUser            = errors.New("not user")
	ErrTextEmpty          = errors.New("text was not provided")
	ErrTextInvalid        = errors.New("type of text is invalid, provide one from string and []styling.StyledTextOption")
	ErrMessageNotExist    = errors.New("message not exist")
	ErrReplyNotMessage    = errors.New("reply header is not a message")
	ErrUnknownTypeMedia   = errors.New("unknown type media")
	ErrUserNotParticipant = errors.New("user not participant")
)

var (
	ErrPeerInfoNotFound        = errors.New("peer info not found in the resolved chats")
	ErrChannelForbidden        = errors.New("peer could not be resolved because channel forbidden")
	ErrContactNotFound         = errors.New("contact not found")
	ErrInvalidBotUsername      = errors.New("provided username was invalid for a bot")
	ErrSendCodeEmptyData       = errors.New("send code returned empty data")
	ErrNotImplemented          = errors.New("not implemented")
	ErrNilSessionStorage       = errors.New("nil session storage is invalid")
	ErrNoAuthenticator         = errors.New("no UserAuthenticator provided")
	ErrInvalidGramjsSession    = errors.New("invalid session string: too short or wrong version")
	ErrInvalidMtcuteSession    = errors.New("invalid mtcute session string: too short or wrong version")
	ErrExtractParticipantUser  = errors.New("could not extract user ID from participant")
	ErrUserNotFoundInPeerStore = errors.New("user not found in peer storage")
)

// Dispatcher control flow errors.
var (
	ErrStopClient       = errors.New("disconnect")
	ErrEndGroups        = errors.New("stopped")
	ErrContinueGroups   = errors.New("continued")
	ErrSkipCurrentGroup = errors.New("skipped")
)
