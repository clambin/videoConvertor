package pipeline

import (
	"context"
	"log/slog"

	"github.com/clambin/xcoder/ffmpeg"
	"golang.org/x/sync/errgroup"
)

func Run(subCtx context.Context, cfg Configuration, queue *Queue, l *slog.Logger) error {
	var g errgroup.Group
	itemCh := make(chan *WorkItem)
	g.Go(func() error { return Scan(subCtx, cfg.Input, queue, itemCh, l.With("component", "scanner")) })
	const inspectorCount = 8
	for range inspectorCount {
		g.Go(func() error {
			Inspect(subCtx, itemCh, cfg, ffmpeg.Probe, fsFileChecker{}, l.With("component", "inspector"))
			return nil
		})
	}
	const converterCount = 2
	for range converterCount {
		g.Go(func() error {
			Transcode(subCtx, queue, cfg, l.With("component", "transcoder"))
			return nil
		})
	}
	return g.Wait()
}
