package main

import (
	"context"
	"flag"
	"github.com/clambin/videoConvertor/internal/server"
	"github.com/clambin/videoConvertor/internal/server/scanner"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

var (
	debug   = flag.Bool("debug", false, "switch on debug logging")
	addr    = flag.String("addr", ":9090", "listener address for API")
	input   = flag.String("input", "/media", "input directory")
	profile = flag.String("profile", "hevc-max", "conversion profile")
	active  = flag.Bool("active", false, "start convertor in active mode")
	remove  = flag.Bool("remove", false, "remove source files after successful conversion")
)

func main() {
	flag.Parse()

	var handlerOpts slog.HandlerOptions
	if *debug {
		handlerOpts.Level = slog.LevelDebug
	}

	l := slog.New(slog.NewTextHandler(os.Stderr, &handlerOpts))

	cfg := server.Config{
		Addr: *addr,
		ScannerConfig: scanner.Config{
			RootDir: *input,
			Profile: *profile,
		},
		RemoveConverted: *remove,
	}
	s, err := server.New(cfg, l)
	if err != nil {
		l.Error("failed to create server", "err", err)
		return
	}
	s.Convertor.SetActive(*active)
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if err = s.Run(ctx); err != nil {
		l.Error("feeder failed", "err", err)
	}
}
