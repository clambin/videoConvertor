package cmd

import (
	"context"
	"fmt"
	"time"

	"codeberg.org/clambin/go-common/charmer"
	"github.com/clambin/xcoder/internal/pipeline"
	"github.com/clambin/xcoder/internal/ui"
	"github.com/rivo/tview"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/sync/errgroup"
)

var (
	runCmd = &cobra.Command{
		Use:   "run",
		Short: "Interactively transcode video files",
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), viper.GetViper())
		},
	}

	runArgs = charmer.Arguments{
		"active":     {Default: false, Help: "start processor in active mode"},
		"input":      {Default: ".", Help: "input directory"},
		"log.format": {Default: "json", Help: "log format"},
		"log.level":  {Default: "info", Help: "log level"},
		"overwrite":  {Default: false, Help: "overwrite existing files"},
		"remove":     {Default: false, Help: "remove source files after successful transcoding"},
		"profile":    {Default: "hevc-high", Help: "transcoding profile"},
	}
)

func init() {
	rootCmd.AddCommand(runCmd)
	if err := charmer.SetPersistentFlags(runCmd, viper.GetViper(), runArgs); err != nil {
		panic(err)
	}
}

func run(ctx context.Context, v *viper.Viper) error {
	cfg, err := pipeline.GetConfigurationFromViper(v)
	if err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	var queue pipeline.Queue
	queue.SetActive(cfg.Active)

	u := ui.New(&queue, cfg)
	l := cfg.Logger(u.LogViewer, nil)
	a := tview.NewApplication().SetRoot(u.Root, true)

	var g errgroup.Group
	subCtx, cancel := context.WithCancel(ctx)
	g.Go(func() error { return pipeline.Run(subCtx, cfg, &queue, l) })
	g.Go(func() error { u.Run(subCtx, a, 250*time.Millisecond); return nil })
	_ = a.Run()

	cancel()
	return g.Wait()
}
