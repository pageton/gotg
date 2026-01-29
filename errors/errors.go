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
