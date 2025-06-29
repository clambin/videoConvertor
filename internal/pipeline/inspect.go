package pipeline

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/clambin/xcoder/ffmpeg"
)

const scanInterval = time.Second

type fileChecker interface {
	TargetIsNewer(a, b string) (bool, error)
}

type Decoder interface {
	Scan(ctx context.Context, path string) (ffmpeg.VideoStats, error)
}

func Inspect(ctx context.Context, ch <-chan *WorkItem, cfg Configuration, probe func(string) (ffmpeg.VideoStats, error), f fileChecker, logger *slog.Logger) {
	ticker := time.NewTicker(scanInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case item := <-ch:
			l := logger.With("source", item.Source)
			err := inspectItem(ctx, item, cfg, probe, f, l)
			var sourceRejectedError *SourceRejectedError
			var sourceSkippedError *SourceSkippedError
			switch {
			case err == nil:
				item.SetWorkStatus(WorkStatus{Status: Inspected})
			case errors.As(err, &sourceRejectedError):
				item.SetWorkStatus(WorkStatus{Status: Rejected, Err: err})
			case errors.As(err, &sourceSkippedError):
				item.SetWorkStatus(WorkStatus{Status: Skipped, Err: err})
			default:
				l.Warn("inspection failed", "err", err)
				item.SetWorkStatus(WorkStatus{Status: Failed, Err: err})
			}
		}
	}
}

func inspectItem(_ context.Context, item *WorkItem, cfg Configuration, probe func(string) (ffmpeg.VideoStats, error), f fileChecker, logger *slog.Logger) (err error) {
	// get the source video stats
	logger.Debug("inspecting file")
	item.Source.VideoStats, err = probe(item.Source.Path)
	if err != nil {
		return fmt.Errorf("probe failed: %w", err)
	}
	logger.Debug("inspection done", "sourceStats", item.Source.VideoStats)

	// Validate that the video meets the criteria and determine how to transcode it
	transcoder, err := cfg.Profile.Inspect(item)
	if err != nil {
		return err
	}

	// check if target already exists and if we're allowed to overwrite it
	targetIsNewer, err := f.TargetIsNewer(item.Source.Path, item.Target.Path)
	if err != nil {
		return err
	}
	if targetIsNewer && !cfg.Overwrite {
		return &SourceSkippedError{Reason: "source is already converted"}
	}

	// Ok to convert
	item.transcoder = transcoder
	return nil
}

type fsFileChecker struct{}

// TargetIsNewer returns true if target exists and is more recent than target. If there was an error checking source, error contains the error.
// If there was an error checking target, it returns (true, nil).
func (c fsFileChecker) TargetIsNewer(source, target string) (bool, error) {
	sStats, err := os.Stat(source)
	if err != nil {
		return false, err
	}
	tStats, err := os.Stat(target)
	if err != nil {
		return false, nil //nolint:nilerr
	}
	return tStats.ModTime().After(sStats.ModTime()), nil
}
