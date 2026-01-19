package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/go-faster/errors"
	"github.com/pageton/gotg"
	"github.com/pageton/gotg/adapter"
	"github.com/pageton/gotg/dispatcher/handlers"
	"github.com/pageton/gotg/dispatcher/handlers/filters"
	"github.com/pageton/gotg/functions"
	"github.com/pageton/gotg/session"
)

func main() {
	appIdEnv := os.Getenv("TG_APP_ID")
	appId, err := strconv.Atoi(appIdEnv)
	if err != nil {
		log.Fatalln("failed to convert app id to int:", err)
	}

	client, err := gotg.NewClient(
		// Get AppID from https://my.telegram.org/apps
		appId,
		// Get ApiHash from https://my.telegram.org/apps
		os.Getenv("TG_API_HASH"),
		// ClientType, as we defined above
		gotg.ClientTypePhone("PHONE_NUMBER_HERE"),
		// Optional parameters of client
		&gotg.ClientOpts{
			InMemory: true,
			Session:  session.SimpleSession(),
		},
	)
	if err != nil {
		log.Fatalln("failed to start client:", err)
	}

	clientDispatcher := client.Dispatcher

	// This Message Handler will download any media passed to bot
	clientDispatcher.AddHandlerToGroup(handlers.NewMessage(filters.Message.Media, download), 1)

	fmt.Printf("client (@%s) has been started...\n", client.Self.Username)

	err = client.Idle()
	if err != nil {
		log.Fatalln("failed to start client:", err)
	}
}

func download(ctx *adapter.Context, update *adapter.Update) error {
	filename, err := functions.GetMediaFileNameWithID(update.EffectiveMessage.Media)
	if err != nil {
		return errors.Wrap(err, "failed to get media file name")
	}

	_, err = ctx.DownloadMedia(
		update.EffectiveMessage.Media,
		adapter.DownloadOutputPath(filename),
		nil,
	)
	if err != nil {
		return errors.Wrap(err, "failed to download media")
	}

	msg := fmt.Sprintf(`File "%s" downloaded`, filename)
	_, err = update.Reply(msg, nil)
	if err != nil {
		return errors.Wrap(err, "failed to reply")
	}

	fmt.Println(msg)

	return nil
}
