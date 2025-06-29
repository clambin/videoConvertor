package pipeline

import (
	"testing"

	"github.com/clambin/xcoder/ffmpeg"
	"github.com/stretchr/testify/assert"
)

func Test_buildTargetFilename(t *testing.T) {
	stats1080 := ffmpeg.VideoStats{Height: 1080}

	tests := []struct {
		name   string
		source string
		stats  ffmpeg.VideoStats
		want   string
	}{
		{
			name:   "valid episode",
			source: "/tmp/foo.S01E01.field.field.field.mp4",
			stats:  stats1080,
			want:   "/directory/foo.s01e01.1080.hevc.mkv",
		},
		{
			name:   "valid episode - no video stats",
			source: "/tmp/foo.S01E01.field.field.field.mp4",
			want:   "/directory/foo.s01e01.hevc.mkv",
		},
		{
			name:   "valid multi-episode",
			source: "/tmp/foo.S01E01E02.field.field.field.mp4",
			stats:  stats1080,
			want:   "/directory/foo.s01e01e02.1080.hevc.mkv",
		},
		{
			name:   "series with spaces",
			source: "/tmp/series 2024 S01E07 attrib1 attrib2 attrib3.mkv",
			stats:  stats1080,
			want:   "/directory/series 2024.s01e07.1080.hevc.mkv",
		},
		{
			name:   "movie (dots)",
			source: "/tmp/foo.bar.2021.field.field.mp4",
			stats:  stats1080,
			want:   "/directory/foo.bar.2021.1080.hevc.mkv",
		},
		{
			name:   "movie (spaces; no video stats)",
			source: "/tmp/foo bar (2021)-field-field.mp4",
			want:   "/directory/foo bar (2021).hevc.mkv",
		},
		{
			name:   "movie (no brackets)",
			source: "/tmp/foo bar snafu 2021 1080p DTS AC3 x264-3Li.mkv",
			stats:  stats1080,
			want:   "/directory/foo bar snafu 2021.1080.hevc.mkv",
		},
		{
			name:   "movie without year",
			source: "/tmp/foo.bar.mp4",
			stats:  stats1080,
			want:   "/directory/foo.bar.1080.hevc.mkv",
		},
		{
			name:   "movie with year",
			source: "/tmp/foo bar (2024).mp4",
			stats:  stats1080,
			want:   "/directory/foo bar (2024).1080.hevc.mkv",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			target := buildTargetFilename(MediaFile{Path: tt.source, VideoStats: tt.stats}, "/directory", "hevc", "mkv")
			assert.Equal(t, tt.want, target)
		})
	}
}
