package main

import (
	"log"

	"github.com/pageton/gotg"
	gorm "github.com/pageton/gotg/ext/gorm"
	web "github.com/pageton/gotg/ext/gorm/examples/auth-using-api-base/web"
	"github.com/pageton/gotg/session"
	"gorm.io/driver/sqlite"
)

func main() {
	wa := web.GetWebAuth()
	// start web api
	go web.Start(wa)
	adapter, err := gorm.New(sqlite.Open("webbot"))
	if err != nil {
		log.Fatalln("failed to create adapter:", err)
	}
	client, err := gotg.NewClient(
		// Get AppID from https://my.telegram.org/apps
		123456,
		// Get ApiHash from https://my.telegram.org/apps
		"API_HASH_HERE",
		// ClientType, as we defined above
		gotg.ClientTypePhone(""),
		// Optional parameters of client
		&gotg.ClientOpts{
			// custom authenticator using web api
			AuthConversator: wa,
			Session:         session.WithAdapter(adapter),
		},
	)
	if err != nil {
		log.Fatalln("failed to start client:", err)
	}
	client.Idle()
}
