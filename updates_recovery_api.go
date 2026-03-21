package gotg

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgerr"
	gotglog "github.com/pageton/gotg/log"
)

const (
	privateChannelSuppressTTL = 6 * time.Hour
	historyFailedSuppressTTL  = 5 * time.Minute
)

type updatesRecoveryAPI interface {
	UpdatesGetState(ctx context.Context) (*tg.UpdatesState, error)
	UpdatesGetDifference(ctx context.Context, request *tg.UpdatesGetDifferenceRequest) (tg.UpdatesDifferenceClass, error)
	UpdatesGetChannelDifference(ctx context.Context, request *tg.UpdatesGetChannelDifferenceRequest) (tg.UpdatesChannelDifferenceClass, error)
}

type updatesRecoveryAPIWrapper struct {
	api    updatesRecoveryAPI
	logger *gotglog.Logger

	mu             sync.Mutex
	suppressedTill map[int64]time.Time
}

func newUpdatesRecoveryAPI(api updatesRecoveryAPI, logger *gotglog.Logger) updatesRecoveryAPI {
	return &updatesRecoveryAPIWrapper{
		api:            api,
		logger:         logger,
		suppressedTill: make(map[int64]time.Time),
	}
}

func (w *updatesRecoveryAPIWrapper) UpdatesGetState(ctx context.Context) (*tg.UpdatesState, error) {
	return w.api.UpdatesGetState(ctx)
}

func (w *updatesRecoveryAPIWrapper) UpdatesGetDifference(ctx context.Context, req *tg.UpdatesGetDifferenceRequest) (tg.UpdatesDifferenceClass, error) {
	return w.api.UpdatesGetDifference(ctx, req)
}

func (w *updatesRecoveryAPIWrapper) UpdatesGetChannelDifference(ctx context.Context, req *tg.UpdatesGetChannelDifferenceRequest) (tg.UpdatesChannelDifferenceClass, error) {
	if req == nil {
		return nil, errors.New("updates.getChannelDifference request is nil")
	}

	channelID := inputChannelID(req.Channel)
	if channelID != 0 {
		if timeout, ok := w.remainingSuppressSeconds(channelID); ok {
			return syntheticChannelDifference(req.Pts, timeout), nil
		}
	}

	diff, err := w.api.UpdatesGetChannelDifference(ctx, req)
	if err == nil {
		return diff, nil
	}

	ttl, shouldSuppress := suppressTTLForError(err)
	if !shouldSuppress {
		return nil, err
	}

	if channelID != 0 {
		w.setSuppressedUntil(channelID, time.Now().Add(ttl))
	}

	if w.logger != nil {
		w.logger.Warn("suppressing updates.getChannelDifference error",
			"channel_id", channelID,
			"error", err,
			"suppress_for", ttl.String(),
		)
	}

	return syntheticChannelDifference(req.Pts, int(ttl.Seconds())), nil
}

func suppressTTLForError(err error) (time.Duration, bool) {
	switch {
	case tgerr.Is(err, "CHANNEL_PRIVATE", "CHANNEL_PUBLIC_GROUP_NA"):
		return privateChannelSuppressTTL, true
	case tgerr.Is(err, "HISTORY_GET_FAILED"):
		return historyFailedSuppressTTL, true
	default:
		return 0, false
	}
}

func inputChannelID(ch tg.InputChannelClass) int64 {
	switch v := ch.(type) {
	case *tg.InputChannel:
		return v.ChannelID
	case interface{ GetChannelID() int64 }:
		return v.GetChannelID()
	default:
		return 0
	}
}

func syntheticChannelDifference(pts int, timeoutSeconds int) *tg.UpdatesChannelDifferenceEmpty {
	diff := &tg.UpdatesChannelDifferenceEmpty{Pts: pts}
	diff.SetFinal(true)
	if timeoutSeconds > 0 {
		diff.SetTimeout(timeoutSeconds)
	}
	return diff
}

func (w *updatesRecoveryAPIWrapper) remainingSuppressSeconds(channelID int64) (int, bool) {
	w.mu.Lock()
	defer w.mu.Unlock()

	until, ok := w.suppressedTill[channelID]
	if !ok {
		return 0, false
	}

	now := time.Now()
	if !now.Before(until) {
		delete(w.suppressedTill, channelID)
		return 0, false
	}

	remaining := int(until.Sub(now).Seconds())
	if remaining < 1 {
		remaining = 1
	}
	return remaining, true
}

func (w *updatesRecoveryAPIWrapper) setSuppressedUntil(channelID int64, until time.Time) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if prev, ok := w.suppressedTill[channelID]; !ok || until.After(prev) {
		w.suppressedTill[channelID] = until
	}
}
