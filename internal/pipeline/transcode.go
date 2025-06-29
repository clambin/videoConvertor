package pipeline

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/clambin/xcoder/ffmpeg"
)

const (
	convertInterval = 100 * time.Millisecond
)

func Transcode(ctx context.Context, queue *Queue, cfg Configuration, logger *slog.Logger) {
	ticker := time.NewTicker(convertInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
		if item := queue.NextToConvert(); item != nil {
			transcodeItem(ctx, item, cfg, logger.With("source", item.Source.Path))
		}
	}
}

func transcodeItem(ctx context.Context, item *WorkItem, cfg Configuration, logger *slog.Logger) {
	// add progress monitor to the transcoder
	tmpDir, err := os.MkdirTemp("", "xcoder")
	if err != nil {
		item.SetWorkStatus(WorkStatus{Status: Failed, Err: err})
		return
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()
	item.transcoder.Progress(progress(item, logger), filepath.Join(tmpDir, "transcoder.sock"))

	// convert the file
	logger.Debug("converting")
	if err = item.transcoder.Run(ctx, logger); err != nil {
		_ = os.Remove(item.Target.Path)
		logger.Warn("conversion failed", "err", err)
		item.SetWorkStatus(WorkStatus{Status: Failed, Err: err})
		return
	}

	// clean up if needed
	if cfg.Remove {
		if err = os.Remove(item.Source.Path); err != nil {
			logger.Warn("failed to remove source file", "err", err)
			item.SetWorkStatus(WorkStatus{Status: Failed, Err: err})
		}
	}
	logger.Info("converted successfully")
	item.SetWorkStatus(WorkStatus{Status: Converted})
}

func progress(item *WorkItem, logger *slog.Logger) func(ffmpeg.Progress) {
	var lastDurationReported time.Duration
	const reportInterval = time.Minute
	item.Progress.Duration = item.Source.VideoStats.Duration

	return func(p ffmpeg.Progress) {
		item.Progress.Update(p)
		if p.Converted-lastDurationReported > reportInterval {
			logger.Info("conversion in progress",
				"progress", p.Converted,
				"speed", p.Speed,
				"completed", strconv.FormatFloat(100*p.Converted.Seconds()/item.Progress.Duration.Seconds(), 'f', 2, 64)+"%",
			)
			lastDurationReported = p.Converted
		}
	}
}
