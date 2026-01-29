package adapter

import (
	"context"
	"io"

	"github.com/gotd/td/telegram/downloader"
	"github.com/gotd/td/tg"
	"github.com/pageton/gotg/functions"
)

// ExportSessionString exports the current session as a string.
func (ctx *Context) ExportSessionString() (string, error) {
	return functions.EncodeSessionToString(ctx.PeerStorage.GetSession())
}

// DownloadOutputClass is an interface for media download destinations.
type DownloadOutputClass interface {
	run(context.Context, *downloader.Builder) (tg.StorageFileTypeClass, error)
}

// DownloadOutputStream downloads media to any io.Writer.
type DownloadOutputStream struct {
	io.Writer
}

func (d DownloadOutputStream) run(ctx context.Context, b *downloader.Builder) (tg.StorageFileTypeClass, error) {
	return b.Stream(ctx, d)
}

// DownloadOutputPath downloads media to a file path.
type DownloadOutputPath string

func (d DownloadOutputPath) run(ctx context.Context, b *downloader.Builder) (tg.StorageFileTypeClass, error) {
	return b.ToPath(ctx, string(d))
}

// DownloadOutputParallel downloads media using parallel chunk downloads.
type DownloadOutputParallel struct {
	io.WriterAt
}

func (d DownloadOutputParallel) run(ctx context.Context, b *downloader.Builder) (tg.StorageFileTypeClass, error) {
	return b.Parallel(ctx, d)
}

// DownloadMediaOpts contains optional parameters for Context.DownloadMedia.
type DownloadMediaOpts struct {
	Threads  int
	Verify   *bool
	PartSize int
}

// DownloadMedia downloads media from a message.
func (ctx *Context) DownloadMedia(media tg.MessageMediaClass, downloadOutput DownloadOutputClass, opts *DownloadMediaOpts) (tg.StorageFileTypeClass, error) {
	if opts == nil {
		opts = &DownloadMediaOpts{}
	}
	mediaDownloader := downloader.NewDownloader()
	if opts.PartSize > 0 {
		mediaDownloader.WithPartSize(opts.PartSize)
	}
	inputFileLocation, err := functions.GetInputFileLocation(media)
	if err != nil {
		return nil, err
	}
	d := mediaDownloader.Download(ctx.Raw, inputFileLocation)
	if opts.Threads > 0 {
		d.WithThreads(opts.Threads)
	}
	if opts.Verify != nil {
		d.WithVerify(*opts.Verify)
	}
	return downloadOutput.run(ctx, d)
}
