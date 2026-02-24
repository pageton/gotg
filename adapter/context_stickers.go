package adapter
import (
	"github.com/gotd/td/tg"
)

func (ctx *Context) GetStickerSet(input tg.InputStickerSetClass) (*tg.MessagesStickerSet, error) {
	result, err := ctx.Raw.MessagesGetStickerSet(ctx.Context, &tg.MessagesGetStickerSetRequest{
		Stickerset: input,
	})
	if err != nil {
		return nil, err
	}
	set, ok := result.(*tg.MessagesStickerSet)
	if !ok {
		return nil, nil
	}
	return set, nil
}

func (ctx *Context) GetStickerSetByShortName(shortName string) (*tg.MessagesStickerSet, error) {
	return ctx.GetStickerSet(&tg.InputStickerSetShortName{ShortName: shortName})
}

func (ctx *Context) CreateStickerSet(req *tg.StickersCreateStickerSetRequest) (*tg.MessagesStickerSet, error) {
	result, err := ctx.Raw.StickersCreateStickerSet(ctx.Context, req)
	if err != nil {
		return nil, err
	}
	set, ok := result.(*tg.MessagesStickerSet)
	if !ok {
		return nil, nil
	}
	return set, nil
}

func (ctx *Context) AddStickerToSet(req *tg.StickersAddStickerToSetRequest) (*tg.MessagesStickerSet, error) {
	result, err := ctx.Raw.StickersAddStickerToSet(ctx.Context, req)
	if err != nil {
		return nil, err
	}
	set, ok := result.(*tg.MessagesStickerSet)
	if !ok {
		return nil, nil
	}
	return set, nil
}

func (ctx *Context) DeleteStickerSet(shortName string) error {
	_, err := ctx.Raw.StickersDeleteStickerSet(ctx.Context, &tg.InputStickerSetShortName{ShortName: shortName})
	return err
}

func (ctx *Context) RemoveStickerFromSet(doc *tg.Document) (*tg.MessagesStickerSet, error) {
	result, err := ctx.Raw.StickersRemoveStickerFromSet(ctx.Context, &tg.InputDocument{
		ID:            doc.ID,
		AccessHash:    doc.AccessHash,
		FileReference: doc.FileReference,
	})
	if err != nil {
		return nil, err
	}
	set, ok := result.(*tg.MessagesStickerSet)
	if !ok {
		return nil, nil
	}
	return set, nil
}