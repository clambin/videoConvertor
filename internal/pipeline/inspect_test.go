package pipeline

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/clambin/videoConvertor/ffmpeg"
	"github.com/clambin/videoConvertor/internal/profile"
	"github.com/stretchr/testify/assert"
)

func TestInspect(t *testing.T) {
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
			name:    "video does not meet criteria",
			profile: "hevc-max",
			args: args{
				stats: ffmpeg.VideoStats{VideoCodec: "h264", Height: 720, BitRate: 1024, Duration: time.Hour, BitsPerSample: 10},
			},
			want: Rejected,
		},
		{
			name:    "video probe failed",
			profile: "hevc-max",
			args: args{
				err: errors.New("failed"),
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
			ff := fakeDecoder{
				stats: tt.args.stats,
				err:   tt.args.err,
			}
			p, _ := profile.GetProfile(tt.profile)

			ch := make(chan *WorkItem)
			l := slog.New(slog.DiscardHandler)
			go func() { Inspect(t.Context(), ch, &ff, p, l) }()

			item := WorkItem{Source: "foo.mkv"}
			ch <- &item
			assert.Eventually(t, func() bool {
				status, _ := item.Status()
				return status == tt.want
			}, time.Second, 10*time.Millisecond)
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
