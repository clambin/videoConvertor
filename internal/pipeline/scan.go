package pipeline

import (
	"context"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"codeberg.org/clambin/go-common/set"
)

var videoExtensions = set.New(".mkv", ".mp4", ".avi", ".mov")

func Scan(ctx context.Context, baseDir string, queue *Queue, ch chan<- *WorkItem, logger *slog.Logger) error {
	return ScanFS(ctx, os.DirFS(baseDir), baseDir, queue, ch, logger)
}

func ScanFS(ctx context.Context, fileSystem fs.FS, baseDir string, queue *Queue, ch chan<- *WorkItem, logger *slog.Logger) error {
	return fs.WalkDir(fileSystem, ".", func(path string, d fs.DirEntry, err error) error {
		select {
		case <-ctx.Done():
			return nil
		default:
		}
		l := logger.With(slog.String("path", path))
		if err != nil {
			l.Warn("failed to scan path", "err", err)
			return nil
		}
		if d.IsDir() {
			l.Debug("not a file")
			return nil
		}
		if !videoExtensions.Contains(strings.ToLower(filepath.Ext(path))) {
			l.Debug("not a video file")
			return nil
		}

		path = filepath.Join(baseDir, path)
		logger.Info("found video file", "path", path)
		ch <- queue.Add(path)
		return nil
	})
}
