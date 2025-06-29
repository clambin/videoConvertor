package ffmpeg

import (
	"context"
	"log/slog"
	"net"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFFMPEG_Build(t *testing.T) {
	tests := []struct {
		name string
		ff   *FFMPEG
		want string
	}{
		{
			name: "simple",
			ff: Decode("foo.mkv").
				Encode(
					"-c:v", "hevc_videotoolbox",
					"-b:v", "1000",
					"-profile:v", "main",
					"-c:a", "copy",
					"-c:s", "copy",
				).
				Muxer("matroska").
				Output("foo.hevc"),
			want: `-i foo.mkv -c:v hevc_videotoolbox -b:v 1000 -profile:v main -c:a copy -c:s copy -f matroska foo.hevc`,
		},
		{
			name: "no output",
			ff: Decode("foo.mkv").
				Encode(
					"-c:v", "hevc_videotoolbox",
					"-profile:v", "main",
					"-c:a", "copy",
					"-c:s", "copy",
				).
				Muxer("matroska"),
			want: `-i foo.mkv -c:v hevc_videotoolbox -profile:v main -c:a copy -c:s copy -f matroska -`,
		},
		{
			name: "full example",
			ff: Decode("foo.mkv", "-hwaccel", "videotoolbox").
				Encode(
					"-c:v", "hevc_videotoolbox",
					"-profile:v", "main",
					"-crf", "10",
					"-c:a", "copy",
					"-c:s", "copy",
				).
				Muxer("matroska").
				Output("foo.hevc").
				LogLevel("error").
				NoStats().
				OverWriteTarget().Progress(func(_ Progress) {}, "/tmp/progress.sock"),
			want: `-hwaccel videotoolbox -i foo.mkv -c:v hevc_videotoolbox -profile:v main -crf 10 -c:a copy -c:s copy -f matroska -loglevel error -nostats -y -progress unix:///tmp/progress.sock foo.hevc`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, strings.Join(tt.ff.Build(t.Context()).Args[1:], " "))
		})
	}
}

func TestFFMPEG_runProgressSocket(t *testing.T) {
	var prog atomic.Value
	ff := FFMPEG{progressSocketPath: t.TempDir() + "/ffmpeg.sock"}
	ff.progress = func(p Progress) {
		prog.Store(p)
	}
	ctx, cancel := context.WithCancel(t.Context())
	go func() {
		require.NoError(t, ff.runProgressSocket(ctx, slog.New(slog.DiscardHandler)))
	}()

	var conn net.Conn
	var err error
	assert.Eventually(t, func() bool {
		conn, err = net.Dial("unix", ff.progressSocketPath)
		return err == nil
	}, time.Second, 50*time.Millisecond)
	t.Cleanup(func() { require.NoError(t, conn.Close()) })

	_, _ = conn.Write([]byte("speed=1.0x\nout_time_ms=1000\nprogress=end\n"))

	assert.Eventually(t, func() bool {
		p, ok := prog.Load().(Progress)
		return ok && p.Speed == 1.0 && p.Converted == time.Millisecond
	}, time.Second, 50*time.Millisecond)
	cancel()
}
