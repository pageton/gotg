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
			Session: sessionMaker.SimpleSession(),
		},
	)
	if err != nil {
		log.Fatalln("failed to start client:", err)
	}

	fmt.Printf("client (@%s) has been started...\n", client.Self.Username)

	client.Idle()
}
