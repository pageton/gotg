package main

import (
	"fmt"
	"log"

	"github.com/pageton/gotg"
	"github.com/pageton/gotg/adapter"
	"github.com/pageton/gotg/dispatcher"
	"github.com/pageton/gotg/dispatcher/handlers"
	"github.com/pageton/gotg/dispatcher/handlers/filters"
	"github.com/pageton/gotg/functions"
	"github.com/pageton/gotg/session"
	"gorm.io/driver/sqlite"
)

func main() {
	client, err := gotg.NewClient(
		123456,
		"API_HASH_HERE",
		gotg.ClientTypePhone("PHONE_NUMBER_HERE"),
		&gotg.ClientOpts{
			Session: session.SqlSession(sqlite.Open("account")),
		},
	)
	if err != nil {
		log.Fatalln("failed to start client:", err)
	}

	dp := client.Dispatcher

	dp.AddHandler(handlers.OnCommand("enable2fa", enable2fa, filters.Private))
	dp.AddHandler(handlers.OnCommand("update2fa", update2fa, filters.Private))
	dp.AddHandler(handlers.OnCommand("disable2fa", disable2fa, filters.Private))
	dp.AddHandler(handlers.OnCommand("sessions", listSessions, filters.Private))
	dp.AddHandler(handlers.OnCommand("revokeall", revokeAllOther, filters.Private))

	fmt.Printf("client (@%s) has been started...\n", client.Self.Username)

	err = client.Idle()
	if err != nil {
		log.Fatalln("client error:", err)
	}
}

func enable2fa(u *adapter.Update) error {
	err := u.Ctx.Enable2FA("my_secure_password", &functions.PasswordOpts{
		Hint:  "the usual one",
		Email: "recovery@example.com",
	})
	if err != nil {
		_, _ = u.Reply(fmt.Sprintf("Failed to enable 2FA: %s", err))
		return dispatcher.EndGroups
	}
	_, _ = u.Reply("2FA has been enabled.")
	return dispatcher.EndGroups
}

func update2fa(u *adapter.Update) error {
	err := u.Ctx.Update2FA("my_secure_password", "my_new_password", &functions.PasswordOpts{
		Hint: "the new one",
	})
	if err != nil {
		_, _ = u.Reply(fmt.Sprintf("Failed to update 2FA: %s", err))
		return dispatcher.EndGroups
	}
	_, _ = u.Reply("2FA password has been updated.")
	return dispatcher.EndGroups
}

func disable2fa(u *adapter.Update) error {
	err := u.Ctx.Disable2FA("my_new_password")
	if err != nil {
		_, _ = u.Reply(fmt.Sprintf("Failed to disable 2FA: %s", err))
		return dispatcher.EndGroups
	}
	_, _ = u.Reply("2FA has been disabled.")
	return dispatcher.EndGroups
}

func listSessions(u *adapter.Update) error {
	sessions, err := u.Ctx.GetActiveSessions()
	if err != nil {
		_, _ = u.Reply(fmt.Sprintf("Failed to get sessions: %s", err))
		return dispatcher.EndGroups
	}
	text := fmt.Sprintf("Active sessions: %d\n\n", len(sessions))
	for _, s := range sessions {
		current := ""
		if s.Current {
			current = " (current)"
		}
		text += fmt.Sprintf("• %s %s — %s %s%s\n",
			s.AppName, s.AppVersion,
			s.DeviceModel, s.Country,
			current,
		)
	}
	_, _ = u.Reply(text)
	return dispatcher.EndGroups
}

func revokeAllOther(u *adapter.Update) error {
	err := u.Ctx.RevokeAllOtherSessions()
	if err != nil {
		_, _ = u.Reply(fmt.Sprintf("Failed to revoke sessions: %s", err))
		return dispatcher.EndGroups
	}
	_, _ = u.Reply("All other sessions have been revoked.")
	return dispatcher.EndGroups
}
