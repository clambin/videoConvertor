package pipeline_test

import (
	"testing"

	"github.com/clambin/xcoder/ffmpeg"
	"github.com/clambin/xcoder/internal/pipeline"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProfile_Inspect(t *testing.T) {
	tests := []struct {
		name   string
		source pipeline.MediaFile
		want   []string
	}{
		{
			name: "hevc 8 bit",
			source: pipeline.MediaFile{
				Path: "foo.mkv",
				VideoStats: ffmpeg.VideoStats{
					VideoCodec:    "h264",
					BitRate:       8_000_000,
					BitsPerSample: 8,
					Height:        1080,
				},
			},
			want: []string{"-hwaccel", "qsv", "-i", "foo.mkv", "-c:v", "hevc_qsv", "-b:v", "4000000", "-profile:v", "main", "-c:a", "copy", "-c:s", "copy", "-f", "matroska", "-nostats", "-loglevel", "error", "foo.1080.hevc.mkv"},
		},
		{
			name: "hevc 10 bit",
			source: pipeline.MediaFile{
				Path: "foo.mkv",
				VideoStats: ffmpeg.VideoStats{
					VideoCodec:    "h264",
					BitRate:       8_000_000,
					BitsPerSample: 10,
					Height:        1080,
				},
			},
			want: []string{"-hwaccel", "qsv", "-i", "foo.mkv", "-c:v", "hevc_qsv", "-b:v", "4000000", "-profile:v", "main10", "-c:a", "copy", "-c:s", "copy", "-f", "matroska", "-nostats", "-loglevel", "error", "foo.1080.hevc.mkv"},
		},
		{
			name: "oversampling",
			source: pipeline.MediaFile{
				Path: "foo.mkv",
				VideoStats: ffmpeg.VideoStats{
					VideoCodec:    "h264",
					BitRate:       16_000_000,
					BitsPerSample: 8,
					Height:        1080,
				},
			},
			want: []string{"-hwaccel", "qsv", "-i", "foo.mkv", "-c:v", "hevc_qsv", "-b:v", "8000000", "-profile:v", "main", "-c:a", "copy", "-c:s", "copy", "-f", "matroska", "-nostats", "-loglevel", "error", "foo.1080.hevc.mkv"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := pipeline.GetProfile("hevc-high")
			require.NoError(t, err)
			xcoder, err := p.Inspect(&pipeline.WorkItem{Source: tt.source})
			require.NoError(t, err)
			cmd := xcoder.Build(t.Context())
			assert.Equal(t, tt.want, cmd.Args[1:])
		})
	}
}
