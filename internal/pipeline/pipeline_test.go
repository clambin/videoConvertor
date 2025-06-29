package pipeline_test

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/clambin/xcoder/internal/pipeline"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRun(t *testing.T) {
	p, _ := pipeline.GetProfile("hevc-high")
	cfg := pipeline.Configuration{
		Profile: p,
		Input:   t.TempDir(),
	}
	var queue pipeline.Queue
	l := slog.New(slog.DiscardHandler)
	var errCh = make(chan error)
	ctx, cancel := context.WithCancel(t.Context())
	go func() { errCh <- pipeline.Run(ctx, cfg, &queue, l) }()

	require.NoError(t, os.WriteFile(filepath.Join(cfg.Input, "video.mkv"), []byte{}, 0644))

	assert.Eventually(t, func() bool {
		items := queue.List()
		if len(queue.List()) != 1 {
			return false
		}
		return items[0].WorkStatus().Status == pipeline.Failed
	}, time.Second, time.Millisecond)

	cancel()
	require.NoError(t, <-errCh)
}
