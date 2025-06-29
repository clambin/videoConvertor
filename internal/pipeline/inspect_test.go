package pipeline

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/clambin/xcoder/ffmpeg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInspect(t *testing.T) {
	// TODO: use the tests from TestProfile_Inspect.
	type args struct {
		stats ffmpeg.VideoStats
		err   error
	}
	tests := []struct {
		name    string
		args    args
		profile string
		want    any
	}{
		{
			name:    "video skipped",
			profile: "hevc-high",
			args: args{
				stats: ffmpeg.VideoStats{VideoCodec: "hevc", Height: 1080, BitRate: 8_000_000, Duration: time.Hour, BitsPerSample: 10},
			},
			want: Skipped,
		},
		{
			name:    "video rejected",
			profile: "hevc-high",
			args: args{
				stats: ffmpeg.VideoStats{VideoCodec: "h264", Height: 720, BitRate: 8_000_000, Duration: time.Hour, BitsPerSample: 10},
			},
			want: Rejected,
		},
		{
			name:    "video probe failed",
			profile: "hevc-high",
			args: args{
				err: errors.New("failed"), //nolint:err113
			},
			want: Failed,
		},
		{
			name:    "video meets criteria",
			profile: "hevc-low",
			args: args{
				stats: ffmpeg.VideoStats{VideoCodec: "h264", Height: 720, BitRate: 1_024_000_000, Duration: time.Hour, BitsPerSample: 10},
			},
			want: Inspected,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := GetProfile(tt.profile)
			require.NoError(t, err)
			ch := make(chan *WorkItem)
			l := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))
			go func() {
				Inspect(t.Context(), ch, Configuration{Profile: p}, func(s string) (ffmpeg.VideoStats, error) {
					return tt.args.stats, tt.args.err
				}, fakeFsChecker{}, l)
			}()

			item := WorkItem{Source: MediaFile{Path: "foo.mkv"}}
			ch <- &item
			assert.Eventually(t, func() bool {
				return item.WorkStatus().Status == tt.want
			}, time.Second, 10*time.Millisecond)
			// t.Log(item.Status())
		})
	}
}

var _ Decoder = &fakeDecoder{}

type fakeDecoder struct {
	stats ffmpeg.VideoStats
	err   error
}

func (f fakeDecoder) Scan(_ context.Context, _ string) (ffmpeg.VideoStats, error) {
	return f.stats, f.err
}
