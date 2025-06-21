package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"time"

	_ "net/http/pprof"

	"github.com/clambin/videoConvertor/internal/configuration"
	"github.com/clambin/videoConvertor/internal/pipeline"
	"github.com/clambin/videoConvertor/internal/processor"
	"github.com/clambin/videoConvertor/internal/ui"
	"github.com/rivo/tview"
	"golang.org/x/sync/errgroup"
)

func main() {
	go func() {
		_ = http.ListenAndServe(":6060", nil)
	}()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	if err := Run(ctx, os.Stderr); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "failed to run. error:", err.Error())
		os.Exit(1)
	}
}

func Run(ctx context.Context, _ io.Writer) error {
	cfg, err := configuration.GetConfiguration()
	if err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	var list pipeline.Queue
	list.SetActive(cfg.Active)

	u := ui.New(&list, cfg)
	l := cfg.Logger(u.LogViewer, nil)
	ff := processor.Processor{Logger: l.With("component", "ffmpeg")}
	a := tview.NewApplication().SetRoot(u.Root, true)
	itemCh := make(chan *pipeline.WorkItem)

	subCtx, cancel := context.WithCancel(ctx)
	var g errgroup.Group
	g.Go(func() error { return pipeline.Scan(subCtx, cfg.Input, &list, itemCh, l.With("component", "scanner")) })
	const inspectorCount = 8
	for range inspectorCount {
		g.Go(func() error {
			pipeline.Inspect(subCtx, itemCh, &ff, cfg.Profile, l.With("component", "inspector"))
			return nil
		})
	}
	const converterCount = 2
	for range converterCount {
		g.Go(func() error { pipeline.Convert(subCtx, &ff, &list, cfg, l.With("component", "processor")); return nil })
	}

	g.Go(func() error { u.Run(subCtx, a, 250*time.Millisecond); return nil })
	_ = a.Run()
	cancel()
	return g.Wait()
}
