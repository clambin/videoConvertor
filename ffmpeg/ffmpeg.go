package ffmpeg

import (
	"bytes"
	"cmp"
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type FFMPEG struct {
	progress           func(Progress)
	output             string
	progressSocketPath string
	args               []string
}

func Decode(path string, args ...string) *FFMPEG {
	f := new(FFMPEG)
	return f.Decode(path, args...)
}

func (ff *FFMPEG) Decode(path string, args ...string) *FFMPEG {
	ff.args = append(ff.args, args...)
	ff.args = append(ff.args, "-i", path)
	return ff
}

func (ff *FFMPEG) Encode(args ...string) *FFMPEG {
	ff.args = append(ff.args, args...)
	return ff
}

func (ff *FFMPEG) Muxer(muxer string, args ...string) *FFMPEG {
	ff.args = append(ff.args, "-f", muxer)
	if len(args) > 0 {
		ff.args = append(ff.args, args...)
	}
	return ff
}

func (ff *FFMPEG) Output(path string) *FFMPEG {
	ff.output = path
	return ff
}

func (ff *FFMPEG) LogLevel(level string) *FFMPEG {
	ff.args = append(ff.args, "-loglevel", level)
	return ff
}

func (ff *FFMPEG) NoStats() *FFMPEG {
	ff.args = append(ff.args, "-nostats")
	return ff
}

func (ff *FFMPEG) OverWriteTarget() *FFMPEG {
	ff.args = append(ff.args, "-y")
	return ff
}

func (ff *FFMPEG) Progress(cb func(Progress), progressSocketPath string) *FFMPEG {
	ff.progress = cb
	ff.progressSocketPath = progressSocketPath
	return ff
}

func (ff *FFMPEG) Build(ctx context.Context) *exec.Cmd {
	args := ff.args
	if ff.progress != nil {
		args = append(args, "-progress", "unix://"+ff.progressSocketPath)
	}
	args = append(args, cmp.Or(ff.output, "-"))
	return exec.CommandContext(ctx, "ffmpeg", args...)
}

func (ff *FFMPEG) Run(ctx context.Context, logger *slog.Logger) error {
	if ff.progressSocketPath != "" {
		if err := ff.runProgressSocket(ctx, logger); err != nil {
			return err
		}
	}
	cmd := ff.Build(ctx)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("ffmpeg: %v (%s)", err, strings.TrimSuffix(stderr.String(), "\n"))
	}
	return err
}

func (ff *FFMPEG) runProgressSocket(ctx context.Context, logger *slog.Logger) error {
	l, err := net.Listen("unix", ff.progressSocketPath)
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}
	go func() {
		if err := serveProgressSocket(ctx, l, ff.progress, logger); err != nil {
			logger.Error("status socket failure", "err", err)
		}
		if err := os.RemoveAll(filepath.Dir(ff.progressSocketPath)); err != nil {
			logger.Error("failed to clean up progress socket", "err", err)
		}
	}()
	return nil
}
