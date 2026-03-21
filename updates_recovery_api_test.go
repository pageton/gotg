package gotg

import (
	"context"
	"errors"
	"testing"

	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgerr"
)

type fakeUpdatesRecoveryAPI struct {
	channelDiffCalls int
	channelDiffFn    func(ctx context.Context, req *tg.UpdatesGetChannelDifferenceRequest) (tg.UpdatesChannelDifferenceClass, error)
}

func (f *fakeUpdatesRecoveryAPI) UpdatesGetState(context.Context) (*tg.UpdatesState, error) {
	return &tg.UpdatesState{}, nil
}

func (f *fakeUpdatesRecoveryAPI) UpdatesGetDifference(context.Context, *tg.UpdatesGetDifferenceRequest) (tg.UpdatesDifferenceClass, error) {
	return &tg.UpdatesDifferenceEmpty{}, nil
}

func (f *fakeUpdatesRecoveryAPI) UpdatesGetChannelDifference(ctx context.Context, req *tg.UpdatesGetChannelDifferenceRequest) (tg.UpdatesChannelDifferenceClass, error) {
	f.channelDiffCalls++
	if f.channelDiffFn != nil {
		return f.channelDiffFn(ctx, req)
	}
	return &tg.UpdatesChannelDifferenceEmpty{Pts: req.Pts}, nil
}

func TestUpdatesRecoveryAPIWrapperSuppressesPrivateChannelErrors(t *testing.T) {
	fake := &fakeUpdatesRecoveryAPI{
		channelDiffFn: func(_ context.Context, _ *tg.UpdatesGetChannelDifferenceRequest) (tg.UpdatesChannelDifferenceClass, error) {
			return nil, tgerr.New(400, "CHANNEL_PRIVATE")
		},
	}
	wrapper := newUpdatesRecoveryAPI(fake, nil)

	req := &tg.UpdatesGetChannelDifferenceRequest{
		Channel: &tg.InputChannel{ChannelID: 123, AccessHash: 456},
		Pts:     77,
		Limit:   100,
		Filter:  &tg.ChannelMessagesFilterEmpty{},
	}

	first, err := wrapper.UpdatesGetChannelDifference(context.Background(), req)
	if err != nil {
		t.Fatalf("first call error: %v", err)
	}
	if fake.channelDiffCalls != 1 {
		t.Fatalf("expected 1 upstream call after first request, got %d", fake.channelDiffCalls)
	}
	empty, ok := first.(*tg.UpdatesChannelDifferenceEmpty)
	if !ok {
		t.Fatalf("expected *tg.UpdatesChannelDifferenceEmpty, got %T", first)
	}
	if empty.Pts != req.Pts {
		t.Fatalf("expected pts=%d, got %d", req.Pts, empty.Pts)
	}

	second, err := wrapper.UpdatesGetChannelDifference(context.Background(), req)
	if err != nil {
		t.Fatalf("second call error: %v", err)
	}
	if fake.channelDiffCalls != 1 {
		t.Fatalf("expected no additional upstream call while suppressed, got %d", fake.channelDiffCalls)
	}
	if _, ok := second.(*tg.UpdatesChannelDifferenceEmpty); !ok {
		t.Fatalf("expected *tg.UpdatesChannelDifferenceEmpty, got %T", second)
	}
}

func TestUpdatesRecoveryAPIWrapperDoesNotSuppressUnknownErrors(t *testing.T) {
	expectedErr := errors.New("boom")
	fake := &fakeUpdatesRecoveryAPI{
		channelDiffFn: func(_ context.Context, _ *tg.UpdatesGetChannelDifferenceRequest) (tg.UpdatesChannelDifferenceClass, error) {
			return nil, expectedErr
		},
	}
	wrapper := newUpdatesRecoveryAPI(fake, nil)

	_, err := wrapper.UpdatesGetChannelDifference(context.Background(), &tg.UpdatesGetChannelDifferenceRequest{
		Channel: &tg.InputChannel{ChannelID: 321, AccessHash: 654},
		Pts:     5,
		Limit:   100,
		Filter:  &tg.ChannelMessagesFilterEmpty{},
	})
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected error %v, got %v", expectedErr, err)
	}
	if fake.channelDiffCalls != 1 {
		t.Fatalf("expected exactly 1 upstream call, got %d", fake.channelDiffCalls)
	}
}

func TestUpdatesRecoveryAPIWrapperSuppressesHistoryGetFailed(t *testing.T) {
	fake := &fakeUpdatesRecoveryAPI{
		channelDiffFn: func(_ context.Context, _ *tg.UpdatesGetChannelDifferenceRequest) (tg.UpdatesChannelDifferenceClass, error) {
			return nil, tgerr.New(400, "HISTORY_GET_FAILED")
		},
	}
	wrapper := newUpdatesRecoveryAPI(fake, nil)

	req := &tg.UpdatesGetChannelDifferenceRequest{
		Channel: &tg.InputChannel{ChannelID: 777, AccessHash: 111},
		Pts:     9,
		Limit:   100,
		Filter:  &tg.ChannelMessagesFilterEmpty{},
	}

	first, err := wrapper.UpdatesGetChannelDifference(context.Background(), req)
	if err != nil {
		t.Fatalf("first call error: %v", err)
	}
	if fake.channelDiffCalls != 1 {
		t.Fatalf("expected 1 upstream call after first request, got %d", fake.channelDiffCalls)
	}
	if _, ok := first.(*tg.UpdatesChannelDifferenceEmpty); !ok {
		t.Fatalf("expected *tg.UpdatesChannelDifferenceEmpty, got %T", first)
	}

	second, err := wrapper.UpdatesGetChannelDifference(context.Background(), req)
	if err != nil {
		t.Fatalf("second call error: %v", err)
	}
	if fake.channelDiffCalls != 1 {
		t.Fatalf("expected no additional upstream call while suppressed, got %d", fake.channelDiffCalls)
	}
	if _, ok := second.(*tg.UpdatesChannelDifferenceEmpty); !ok {
		t.Fatalf("expected *tg.UpdatesChannelDifferenceEmpty, got %T", second)
	}
}

func TestUpdatesRecoveryAPIWrapperNilRequest(t *testing.T) {
	fake := &fakeUpdatesRecoveryAPI{}
	wrapper := newUpdatesRecoveryAPI(fake, nil)

	_, err := wrapper.UpdatesGetChannelDifference(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error for nil request")
	}
	if fake.channelDiffCalls != 0 {
		t.Fatalf("expected 0 upstream calls for nil request, got %d", fake.channelDiffCalls)
	}
}
