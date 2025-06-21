package pipeline

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/clambin/videoConvertor/ffmpeg"
	"github.com/clambin/videoConvertor/internal/profile"
)

type Decoder interface {
	Scan(ctx context.Context, path string) (ffmpeg.VideoStats, error)
}

const scanInterval = time.Second

func Inspect(ctx context.Context, ch <-chan *WorkItem, decoder Decoder, profile profile.Profile, logger *slog.Logger) {
	ticker := time.NewTicker(scanInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case item := <-ch:
			l := logger.With("source", item.Source)
			if err := inspectItem(ctx, item, decoder, profile, l); err != nil {
				l.Warn("preprocessing failed", "err", err)
			}
		}
	}
}

func inspectItem(ctx context.Context, item *WorkItem, codec Decoder, videoProfile profile.Profile, l *slog.Logger) error {
	l.Debug("inspecting file")
	sourceStats, err := codec.Scan(ctx, item.Source)
	if err != nil {
		l.Warn("failed to scan file", "err", err)
		err = fmt.Errorf("inspection failed: %w", err)
		item.SetStatus(Failed, err)
		return err
	}
	item.AddSourceStats(sourceStats)
	l.Debug("inspection done", "sourceStats", sourceStats)

	// Validate that the video meets the criteria
	targetStats, err := videoProfile.Evaluate(sourceStats)
	if err != nil {
		l.Debug("should not convert file", "err", err)
		status := Rejected
		if errors.Is(err, profile.ErrSourceInTargetCodec) {
			status = Skipped
		}
		item.SetStatus(status, err)
		return err
	}

	// Ok to convert
	item.AddTargetStats(targetStats)
	item.SetStatus(Inspected, nil)
	return nil
}
