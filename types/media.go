package types

import (
	"fmt"

	"github.com/gotd/td/fileid"
	"github.com/gotd/td/tg"
)

// Document wraps tg.Document with helper methods.
// It provides a FileID() method that returns a base64url encoded file identifier
// compatible with the Bot API format.
type Document struct {
	*tg.Document
}

// NewDocument creates a new Document wrapper from a tg.Document.
// Returns nil if the input document is nil.
func NewDocument(doc *tg.Document) *Document {
	if doc == nil {
		return nil
	}
	return &Document{Document: doc}
}

// FileID returns an encoded Bot API file ID string for this document.
// Uses the official github.com/gotd/td/fileid package for encoding.
// Returns empty string if the document is nil.
//
// Example:
//
//	doc := msg.Document()
//	if doc != nil {
//	    fileID := doc.FileID()
//	    if fileID != "" {
//	        fmt.Printf("Document FileID: %s\n", fileID)
//	    }
//	}
func (d *Document) FileID() string {
	if d == nil || d.Document == nil {
		return ""
	}
	fid := fileid.FromDocument(d.Document)
	encoded, err := fileid.EncodeFileID(fid)
	if err != nil {
		return ""
	}
	return encoded
}

// Photo wraps tg.Photo with helper methods.
// It provides a FileID() method that returns a base64url encoded file identifier
// compatible with the Bot API format.
type Photo struct {
	*tg.Photo
}

// NewPhoto creates a new Photo wrapper from a tg.Photo.
// Returns nil if the input photo is nil.
func NewPhoto(photo *tg.Photo) *Photo {
	if photo == nil {
		return nil
	}
	return &Photo{Photo: photo}
}

// FileID returns an encoded Bot API file ID string for this photo.
// Uses the official github.com/gotd/td/fileid package for encoding.
// Returns empty string if the photo is nil.
//
// Example:
//
//	photo := msg.Photo()
//	if photo != nil {
//	    fileID := photo.FileID()
//	    if fileID != "" {
//	        fmt.Printf("Photo FileID: %s\n", fileID)
//	    }
//	}
func (p *Photo) FileID() string {
	if p == nil || p.Photo == nil {
		return ""
	}
	fid := fileid.FromPhoto(p.Photo, 0) // 0 for full photo
	encoded, err := fileid.EncodeFileID(fid)
	if err != nil {
		return ""
	}
	return encoded
}

// InputMediaFromFileID creates tg.InputMediaClass from an encoded fileID string.
// Uses the official github.com/gotd/td/fileid package for decoding.
// Returns an error if the fileID is invalid or the file type is unsupported.
//
// The fileID should be obtained from a previous Message's FileID(), Document().FileID(), or Photo().FileID() method.
//
// Example:
//
//	fileID := msg.FileID()
//	inputMedia, err := types.InputMediaFromFileID(fileID, "Check this out!")
//	if err != nil {
//	    return err
//	}
//	ctx.SendMedia(chatID, &tg.MessagesSendMediaRequest{
//	    Media: inputMedia,
//	})
func InputMediaFromFileID(fileIDStr string, caption string) (tg.InputMediaClass, error) {
	fid, err := fileid.DecodeFileID(fileIDStr)
	if err != nil {
		return nil, fmt.Errorf("decode fileID: %w", err)
	}

	switch fid.Type {
	case fileid.Document, fileid.Video, fileid.Audio, fileid.Animation,
		fileid.Voice, fileid.VideoNote, fileid.Sticker:
		return &tg.InputMediaDocument{
			ID: &tg.InputDocument{
				ID:            fid.ID,
				AccessHash:    fid.AccessHash,
				FileReference: fid.FileReference,
			},
		}, nil
	case fileid.Photo:
		return &tg.InputMediaPhoto{
			ID: &tg.InputPhoto{
				ID:            fid.ID,
				AccessHash:    fid.AccessHash,
				FileReference: fid.FileReference,
			},
		}, nil
	default:
		return nil, fmt.Errorf("unsupported file type: %v", fid.Type)
	}
}

