package main

import (
	"fmt"
	"log"

	"github.com/pageton/gotg"
	"github.com/pageton/gotg/sessionMaker"
)

func main() {
	client, err := gotg.NewClient(
		// Get AppID from https://my.telegram.org/apps
		123456,
		// Get ApiHash from https://my.telegram.org/apps
		"API_HASH_HERE",
		// ClientType, as we defined above
		gotg.ClientTypePhone("PHONE_NUMBER_HERE"),
		// Optional parameters of client
		&gotg.ClientOpts{
			Session: sessionMaker.GramjsSession("enter session string here").
				// Sqlite session name (if you're not using memory session)
				// i.e. InMemory in ClientOpts is set to false
				// It will be saved as my_session.session as per this example.
				Name("my_session"),
		},
	)
	if err != nil {
		log.Fatalln("failed to start client:", err)
	}

	fmt.Printf("client (@%s) has been started...\n", client.Self.Username)

	client.Idle()
}
