package filters

import "github.com/gotd/td/tg"

type businessDeletedMsg struct{}

func (*businessDeletedMsg) All(_ *tg.UpdateBotDeleteBusinessMessage) bool {
	return true
}

func (*businessDeletedMsg) ConnectionID(id string) BusinessDeletedMessageFilter {
	return func(bd *tg.UpdateBotDeleteBusinessMessage) bool {
		return bd.ConnectionID == id
	}
}
