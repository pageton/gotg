package functions

import (
	"context"

	"github.com/gotd/td/tg"
	"github.com/pageton/gotg/errors"
	"github.com/pageton/gotg/storage"
)

// ArchiveChats moves chats to the archive folder (folder ID 1).
//
// Example:
//
//	peers := []tg.InputPeerClass{chatPeer1, chatPeer2}
//	success, err := functions.ArchiveChats(ctx, client.Raw, peers)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// Parameters:
//   - ctx: Context for the API call
//   - client: The raw Telegram client
//   - peers: List of peers to archive
//
// Returns true if successful, or an error.
func ArchiveChats(ctx context.Context, client *tg.Client, peers []tg.InputPeerClass) (bool, error) {
	var folderPeers = make([]tg.InputFolderPeer, len(peers))
	for n, peer := range peers {
		folderPeers[n] = tg.InputFolderPeer{
			Peer:     peer,
			FolderID: 1,
		}
	}
	_, err := client.FoldersEditPeerFolders(ctx, folderPeers)
	return err == nil, err
}

// UnarchiveChats moves chats out of the archive folder (folder ID 0).
//
// Example:
//
//	peers := []tg.InputPeerClass{chatPeer1, chatPeer2}
//	success, err := functions.UnarchiveChats(ctx, client.Raw, peers)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// Parameters:
//   - ctx: Context for the API call
//   - client: The raw Telegram client
//   - peers: List of peers to unarchive
//
// Returns true if successful, or an error.
func UnarchiveChats(ctx context.Context, client *tg.Client, peers []tg.InputPeerClass) (bool, error) {
	var folderPeers = make([]tg.InputFolderPeer, len(peers))
	for n, peer := range peers {
		folderPeers[n] = tg.InputFolderPeer{
			Peer:     peer,
			FolderID: 0,
		}
	}
	_, err := client.FoldersEditPeerFolders(ctx, folderPeers)
	return err == nil, err
}

// GetChatInviteLink generates an invite link for a chat.
//
// Parameters:
//   - ctx: Context for API call
//   - raw: The raw Telegram client
//   - p: Peer storage for resolving peer references
//   - chatID: The chat ID to generate invite link for
//   - req: Telegram's MessagesExportChatInviteRequest (use &tg.MessagesExportChatInviteRequest{} for default)
//
// Returns exported chat invite or an error.
func GetChatInviteLink(ctx context.Context, raw *tg.Client, p *storage.PeerStorage, chatID int64, req ...*tg.MessagesExportChatInviteRequest) (tg.ExportedChatInviteClass, error) {
	inputPeer := GetInputPeerClassFromID(p, chatID)
	if inputPeer == nil {
		return nil, errors.ErrPeerNotFound
	}

	optReq := GetOptDef(&tg.MessagesExportChatInviteRequest{}, req...)

	switch inputPeer.(type) {
	case *tg.InputPeerChannel, *tg.InputPeerChat:
		link, err := raw.MessagesExportChatInvite(ctx, &tg.MessagesExportChatInviteRequest{
			Peer:                  inputPeer,
			LegacyRevokePermanent: optReq.LegacyRevokePermanent,
			RequestNeeded:         optReq.RequestNeeded,
			UsageLimit:            optReq.UsageLimit,
			Title:                 optReq.Title,
			ExpireDate:            optReq.ExpireDate,
		})
		if err != nil {
			return nil, err
		}
		return link, nil
	default:
		return nil, errors.ErrNotChat
	}
}
