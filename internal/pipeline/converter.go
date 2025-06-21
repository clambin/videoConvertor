package pipeline

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/clambin/videoConvertor/internal/configuration"
	"github.com/clambin/videoConvertor/internal/processor"
)

const (
	convertInterval = 100 * time.Millisecond
)

func Convert(ctx context.Context, codec Codec, queue *Queue, cfg configuration.Configuration, logger *slog.Logger) {
	convertWithFileChecker(ctx, codec, queue, fsFileChecker{}, cfg, logger)
}

type Codec interface {
	Convert(ctx context.Context, request processor.Request) error
}

type fileChecker interface {
	TargetIsNewer(a, b string) (bool, error)
}

func convertWithFileChecker(ctx context.Context, codec Codec, queue *Queue, f fileChecker, cfg configuration.Configuration, logger *slog.Logger) {
	ticker := time.NewTicker(convertInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
		if item := queue.NextToConvert(); item != nil {
			convertItem(ctx, item, codec, f, cfg, logger)
		}
	}
}

func convertItem(ctx context.Context, item *WorkItem, codec Codec, f fileChecker, cfg configuration.Configuration, logger *slog.Logger) {
	// Build target name
	target := buildTargetFilename(item, "", cfg.Profile.Codec, "mkv")
	logger = logger.With("target", target)

	// Has the file already been converted?
	targetIsNewer, err := f.TargetIsNewer(item.Source, target)
	if err != nil {
		logger.Error("failed to check if target is newer", "err", err)
		item.SetStatus(Failed, err)
	}
	if targetIsNewer && !cfg.Overwrite {
		logger.Info("already converted")
		item.SetStatus(Skipped, err)
		return
	}

	// build the request
	req := processor.Request{
		Source:      item.Source,
		Target:      target,
		TargetStats: item.TargetVideoStats(),
	}

	logger.Info("target determined", "bitrate", req.TargetStats.BitRate)

	var lastDurationReported time.Duration
	const reportInterval = 1 * time.Minute
	totalDuration := item.SourceVideoStats().Duration
	req.ProgressCB = func(progress processor.Progress) {
		completed := progress.Converted.Seconds() / totalDuration.Seconds()
		item.Progress.Update(progress)
		if progress.Converted-lastDurationReported > reportInterval {
			logger.Info("conversion in progress", "progress", progress.Converted, "speed", progress.Speed, "completed", completed)
			lastDurationReported = progress.Converted
		}
	}

	// convert the file
	logger.Debug("converting")
	if err = codec.Convert(ctx, req); err != nil {
		_ = os.Remove(target)
		logger.Warn("conversion failed", "err", err)
		item.SetStatus(Failed, err)
		return
	}

	// clean up if needed
	if cfg.Remove {
		if err = os.Remove(item.Source); err != nil {
			logger.Warn("failed to remove source file", "err", err)
			item.SetStatus(Failed, err)
		}
	}
	logger.Info("converted successfully")
	item.SetStatus(Converted, nil)
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
		return false, nil
	}
	return tStats.ModTime().After(sStats.ModTime()), nil
}
