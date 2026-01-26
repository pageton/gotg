package types

import (
	"fmt"

	"github.com/gotd/td/telegram/downloader"
	"github.com/gotd/td/tg"
	"github.com/pageton/gotg/functions"
)

// Download downloads the media from this message to a file path.
// If path is empty, auto-generates a filename using GetMediaFileNameWithID.
// Returns the file path that was used and the file type.
//
// Example:
//
//	path, fileType, err := msg.Download("")
//	path, fileType, err := msg.Download("downloads/photo.jpg")
func (m *Message) Download(path string) (string, tg.StorageFileTypeClass, error) {
	if m.RawClient == nil {
		return "", nil, fmt.Errorf("message has no client context")
	}
	if m.Media == nil {
		return "", nil, fmt.Errorf("message has no media")
	}

	if path == "" {
		var err error
		path, err = functions.GetMediaFileNameWithID(m.Media)
		if err != nil {
			return "", nil, err
		}
	}

	inputFileLocation, err := functions.GetInputFileLocation(m.Media)
	if err != nil {
		return "", nil, err
	}

	mediaDownloader := downloader.NewDownloader()
	d := mediaDownloader.Download(m.RawClient, inputFileLocation)

	fileType, err := d.ToPath(m.Ctx, path)
	if err != nil {
		return "", nil, err
	}

	return path, fileType, nil
}

// DownloadBytes downloads the media from this message into memory.
// Returns the file content as bytes and the file type.
//
// Example:
//
//	data, fileType, err := msg.DownloadBytes()
func (m *Message) DownloadBytes() ([]byte, tg.StorageFileTypeClass, error) {
	if m.RawClient == nil {
		return nil, nil, fmt.Errorf("message has no client context")
	}
	if m.Media == nil {
		return nil, nil, fmt.Errorf("message has no media")
	}

	inputFileLocation, err := functions.GetInputFileLocation(m.Media)
	if err != nil {
		return nil, nil, err
	}

	mediaDownloader := downloader.NewDownloader()
	d := mediaDownloader.Download(m.RawClient, inputFileLocation)

	var buf []byte
	fileType, err := d.Stream(m.Ctx, &byteWriter{b: &buf})
	if err != nil {
		return nil, nil, err
	}

	return buf, fileType, nil
}

type byteWriter struct {
	b *[]byte
}

func (bw *byteWriter) Write(p []byte) (n int, err error) {
	*bw.b = append(*bw.b, p...)
	return len(p), nil
}
