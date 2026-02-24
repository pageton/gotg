package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/gotd/td/session/tdesktop"
	"github.com/pageton/gotg"
	"github.com/pageton/gotg/session"
)

func main() {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatalln(err)
	}
	telegramDir := filepath.Join(home, ".local/share/TelegramDesktop")
	accounts, err := tdesktop.Read(telegramDir, nil)
	if err != nil {
		log.Fatalln(err)
	}

	client, err := gotg.NewClient(
		// Get AppID from https://my.telegram.org/apps
		123456,
		// Get ApiHash from https://my.telegram.org/apps
		"API_HASH_HERE",
		// ClientType, as we defined above
		gotg.ClientTypePhone("PHONE_NUMBER_HERE"),
		// Optional parameters of client
		&gotg.ClientOpts{
			// There can be up to 3 tdesktop.Account, we consider here there is
			// at least a single on, you can loop through them with
			// for _, account := range accounts {// your code}
			Session: session.TdataSession(accounts[0]).Name("tdata"),
		},
	)
	if err != nil {
		log.Fatalln("failed to start client:", err)
	}
	fmt.Printf("client (@%s) has been started...\n", client.Self.Username)

	client.Idle()
}
