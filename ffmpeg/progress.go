package ffmpeg

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net"
	"strconv"
	"time"
)

// serveProgressSocket reads the ffmpeg progress and calls the process callback function.
func serveProgressSocket(ctx context.Context, l net.Listener, progressCallback func(Progress), logger *slog.Logger) error {
	fd, err := l.Accept()
	if err != nil {
		return fmt.Errorf("failed to serve status socket: %w", err)
	}

	defer func() { _ = fd.Close() }()

	ch := progress(fd, logger)
	for {
		select {
		case prog, ok := <-ch:
			if !ok {
				logger.Debug("progress socket closed")
				return nil
			}
			progressCallback(prog)
		case <-ctx.Done():
			logger.Debug("context cancelled", "err", ctx.Err())
			return nil
		}
	}
}

type Progress struct {
	Converted time.Duration
	Speed     float64
}

// progress reads the ffmpeg progress information and returns Progress records on a channel.
func progress(r io.Reader, logger *slog.Logger) chan Progress {
	var (
		convertedMarker = []byte("out_time_ms=")
		speedMarker     = []byte("speed=")
		endMarker       = []byte("progress=end")
	)
	ch := make(chan Progress)
	go func() {
		defer close(ch)
		s := bufio.NewScanner(r)
		var haveProgress, haveSpeed bool
		var prog Progress
		for s.Scan() {
			line := s.Bytes()
			switch {
			case bytes.HasPrefix(line, convertedMarker):
				microSeconds, _ := strconv.Atoi(string(line[len(convertedMarker):]))
				prog.Converted = time.Duration(microSeconds) * time.Microsecond
				haveProgress = true
			case bytes.HasPrefix(line, speedMarker):
				line = bytes.TrimSuffix(line, []byte("x"))
				prog.Speed, _ = strconv.ParseFloat(string(line[len(speedMarker):]), 64)
				haveSpeed = true
			}
			if haveProgress && haveSpeed {
				ch <- prog
				haveProgress = false
				haveSpeed = false
			}
			if bytes.Equal(line, endMarker) {
				return
			}
		}
		if err := s.Err(); err != nil {
			logger.Error("failed to read progress socket", "err", err)
		}
	}()
	return ch
}
