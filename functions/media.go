package functions

import (
	"context"
	"fmt"

	"github.com/gotd/td/tg"
	"github.com/pageton/gotg/errors"
	"github.com/pageton/gotg/storage"
)

// GetMediaFileNameWithID returns media's filename in format "{id}-{name}.{extension}".
//
// Parameters:
//   - media: The media object to get filename from
//
// Returns formatted filename or an error.
func GetMediaFileNameWithID(media tg.MessageMediaClass) (string, error) {
	switch v := media.(type) {
	case *tg.MessageMediaPhoto:
		f, ok := v.Photo.AsNotEmpty()
		if !ok {
			return "", errors.ErrUnknownTypeMedia
		}
		return fmt.Sprintf("%d.png", f.ID), nil
	case *tg.MessageMediaDocument:
		var (
			attr             tg.DocumentAttributeClass
			ok               bool
			filenameFromAttr *tg.DocumentAttributeFilename
			f                *tg.Document
			filename         = "undefined"
		)
		f, ok = v.Document.AsNotEmpty()
		if !ok {
			return "", errors.ErrUnknownTypeMedia
		}
		for _, attr = range f.Attributes {
			filenameFromAttr, ok = attr.(*tg.DocumentAttributeFilename)
			if ok {
				filename = filenameFromAttr.FileName
			}
		}
		return fmt.Sprintf("%d-%s", f.ID, filename), nil
	case *tg.MessageMediaStory:
		f, ok := v.Story.(*tg.StoryItem)
		if !ok {
			return "", errors.ErrUnknownTypeMedia
		}
		return GetMediaFileNameWithID(f.Media)
	}
	return "", errors.ErrUnknownTypeMedia
}

// GetMediaFileName returns media's filename in format "{name}.{extension}".
//
// Parameters:
//   - media: The media object to get filename from
//
// Returns filename or an error.
func GetMediaFileName(media tg.MessageMediaClass) (string, error) {
	switch v := media.(type) {
	case *tg.MessageMediaPhoto:
		f, ok := v.Photo.AsNotEmpty()
		if !ok {
			return "", errors.ErrUnknownTypeMedia
		}
		return fmt.Sprintf("%d.png", f.ID), nil
	case *tg.MessageMediaDocument:
		var (
			attr             tg.DocumentAttributeClass
			ok               bool
			filenameFromAttr *tg.DocumentAttributeFilename
			f                *tg.Document
			filename         = "undefined"
		)
		f, ok = v.Document.AsNotEmpty()
		if !ok {
			return "", errors.ErrUnknownTypeMedia
		}
		for _, attr = range f.Attributes {
			filenameFromAttr, ok = attr.(*tg.DocumentAttributeFilename)
			if ok {
				filename = filenameFromAttr.FileName
			}
		}
		return filename, nil
	case *tg.MessageMediaStory:
		f, ok := v.Story.(*tg.StoryItem)
		if !ok {
			return "", errors.ErrUnknownTypeMedia
		}
		return GetMediaFileName(f.Media)
	}
	return "", errors.ErrUnknownTypeMedia
}

// GetInputFileLocation returns tg.InputFileLocationClass, which can be used to download media.
//
// Parameters:
//   - media: The media object to get location for
//
// Returns input file location or an error.
func GetInputFileLocation(media tg.MessageMediaClass) (tg.InputFileLocationClass, error) {
	switch v := media.(type) {
	case *tg.MessageMediaPhoto:
		f, ok := v.Photo.AsNotEmpty()
		if !ok {
			return nil, errors.ErrUnknownTypeMedia
		}
		thumbSize := ""
		if len(f.Sizes) > 1 {
			// Lowest (f.Sizes[0]) size has the lowest resolution
			// Highest (f.Sizes[len(f.Sizes)-1]) has the highest resolution
			thumbSize = f.Sizes[len(f.Sizes)-1].GetType()
		}
		return &tg.InputPhotoFileLocation{
			ID:            f.ID,
			AccessHash:    f.AccessHash,
			FileReference: f.FileReference,
			ThumbSize:     thumbSize,
		}, nil
	case *tg.MessageMediaDocument:
		f, ok := v.Document.AsNotEmpty()
		if !ok {
			return nil, errors.ErrUnknownTypeMedia
		}
		return f.AsInputDocumentFileLocation(), nil
	case *tg.MessageMediaStory:
		f, ok := v.Story.(*tg.StoryItem)
		if !ok {
			return nil, errors.ErrUnknownTypeMedia
		}
		return GetInputFileLocation(f.Media)
	}
	return nil, errors.ErrUnknownTypeMedia
}

// GetUserProfilePhotos fetches photos from a user's profile.
//
// Parameters:
//   - ctx: Context for the API call
//   - raw: The raw Telegram client
//   - p: Peer storage for resolving peer references
//   - userID: The user ID to get photos from
//   - opts: Optional parameters for photos request
//
// Returns list of photos or an error.
func GetUserProfilePhotos(ctx context.Context, raw *tg.Client, p *storage.PeerStorage, userID int64, opts *tg.PhotosGetUserPhotosRequest) ([]tg.PhotoClass, error) {
	peerUser := GetInputPeerClassFromID(p, userID)
	if peerUser == nil {
		return nil, errors.ErrPeerNotFound
	}

	if opts == nil {
		opts = &tg.PhotosGetUserPhotosRequest{}
	}

	switch peer := peerUser.(type) {
	case *tg.InputPeerUser:
		opts.UserID = &tg.InputUser{
			UserID:     peer.UserID,
			AccessHash: peer.AccessHash,
		}
	default:
		return nil, errors.ErrNotUser
	}

	photosResult, err := raw.PhotosGetUserPhotos(ctx, opts)
	if err != nil {
		return nil, err
	}

	switch result := photosResult.(type) {
	case *tg.PhotosPhotos:
		return result.Photos, nil
	case *tg.PhotosPhotosSlice:
		return result.Photos, nil
	default:
		return nil, errors.ErrUnknownTypeMedia
	}
}

// TransferStarGift transfers a star gift to a chat.
//
// Parameters:
//   - ctx: Context for the API call
//   - raw: The raw Telegram client
//   - p: Peer storage for resolving peer references
//   - chatID: The chat ID to send gift to
//   - starGift: The star gift to transfer
//
// Returns updates confirming the action or an error.
func TransferStarGift(ctx context.Context, raw *tg.Client, p *storage.PeerStorage, chatID int64, starGift tg.InputSavedStarGiftClass) (tg.UpdatesClass, error) {
	peerUser := GetInputPeerClassFromID(p, chatID)
	if peerUser == nil {
		return nil, errors.ErrPeerNotFound
	}

	upd, err := raw.PaymentsTransferStarGift(ctx, &tg.PaymentsTransferStarGiftRequest{
		ToID:     peerUser,
		Stargift: starGift,
	})
	if err != nil {
		return nil, err
	}

	return upd, nil
}
