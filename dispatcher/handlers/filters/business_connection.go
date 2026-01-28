package filters

import "github.com/gotd/td/tg"

type businessConnection struct{}

func (*businessConnection) All(_ *tg.UpdateBotBusinessConnect) bool {
	return true
}

func (*businessConnection) Enabled(bc *tg.UpdateBotBusinessConnect) bool {
	return !bc.Connection.Disabled
}

func (*businessConnection) Disabled(bc *tg.UpdateBotBusinessConnect) bool {
	return bc.Connection.Disabled
}

func (*businessConnection) UserID(userID int64) BusinessConnectionFilter {
	return func(bc *tg.UpdateBotBusinessConnect) bool {
		return bc.Connection.UserID == userID
	}
}

func (*businessConnection) ConnectionID(id string) BusinessConnectionFilter {
	return func(bc *tg.UpdateBotBusinessConnect) bool {
		return bc.Connection.ConnectionID == id
	}
}
