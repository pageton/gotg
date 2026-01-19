package main

import (
	"fmt"
	"log"

	"github.com/glebarez/sqlite"
	"github.com/go-faster/errors"
	"github.com/pageton/gotg"
	"github.com/pageton/gotg/dispatcher/handlers"
	"github.com/pageton/gotg/dispatcher/handlers/filters"
	"github.com/pageton/gotg/adapter"
	"github.com/pageton/gotg/session"
)

func main() {
	client, err := gotg.NewClient(
		123456,
		"API_HASH_HERE",
		gotg.AsBot("BOT_TOKEN_HERE"),
		&gotg.ClientOpts{
			Session: session.SqlSession(sqlite.Open("dlbot")),
		},
	)
	if err != nil {
		log.Fatalln("failed to start client:", err)
	}

	dp := client.Dispatcher

	// Download command - downloads the replied message or any media message
	dp.AddHandler(handlers.OnCommand("download", downloadCommand))
	dp.AddHandler(handlers.OnCommand("download_bytes", downloadBytesCommand))

	// Auto-download any media sent to the bot
	dp.AddHandler(handlers.OnMessage(autoDownload, filters.Message.Media))

	fmt.Printf("Bot (@%s) has been started...\n", client.Self.Username)

	client.Idle()
}

// downloadCommand downloads the replied message's media to disk
func downloadCommand(u *adapter.Update) error {
	// Get the reply message using EffectiveReply
	reply := u.EffectiveReply()
	if reply == nil {
		_, err := u.Reply("Please reply to a media message to download it.", nil)
		return err
	}

	if reply.Media == nil {
		_, err := u.Reply("The replied message has no media.", nil)
		return err
	}

	// Download to auto-generated filename
	path, fileType, err := reply.Download("")
	if err != nil {
		_, err := u.Reply(fmt.Sprintf("Failed to download: %v", err), nil)
		return err
	}

	msg := fmt.Sprintf("✅ Downloaded to: %s\nFile type: %T", path, fileType)
	_, err = u.Reply(msg, nil)
	return err
}

// downloadBytesCommand downloads the replied message's media into memory
func downloadBytesCommand(u *adapter.Update) error {
	// Get the reply message using EffectiveReply
	reply := u.EffectiveReply()
	if reply == nil {
		_, err := u.Reply("Please reply to a media message to download it.", nil)
		return err
	}

	if reply.Media == nil {
		_, err := u.Reply("The replied message has no media.", nil)
		return err
	}

	// Download into memory
	data, fileType, err := reply.DownloadBytes()
	if err != nil {
		_, err := u.Reply(fmt.Sprintf("Failed to download: %v", err), nil)
		return err
	}

	msg := fmt.Sprintf("✅ Downloaded %d bytes into memory\nFile type: %T", len(data), fileType)
	_, err = u.Reply(msg, nil)
	return err
}

// autoDownload automatically downloads any media sent to the bot
func autoDownload(u *adapter.Update) error {
	msg := u.EffectiveMessage

	// Download the message's media directly
	path, fileType, err := msg.Download("downloads")
	if err != nil {
		return errors.Wrap(err, "failed to download media")
	}

	replyMsg := fmt.Sprintf("📥 Auto-downloaded to: %s\nFile type: %T", path, fileType)
	_, err = u.Reply(replyMsg, nil)
	return err
}
