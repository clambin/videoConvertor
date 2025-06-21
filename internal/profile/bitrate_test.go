package profile

import (
	"testing"

	"github.com/clambin/videoConvertor/ffmpeg"
	"github.com/stretchr/testify/assert"
)

func Test_bitRates_getBitrate(t *testing.T) {
	b := bitRates{{height: 100, bitrate: 1000}, {height: 200, bitrate: 2000}, {height: 300, bitrate: 3000}}
	tests := []struct {
		name   string
		height int
		want   int
	}{
		{"too low", 50, 1000},
		{"match low", 200, 2000},
		{"interpolate", 250, 2500},
		{"match high", 300, 3000},
		{"too high", 400, 3000},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, b.getBitrate(tt.height))
		})
	}
}

func Test_getTargetBitRate(t *testing.T) {
	tests := []struct {
		name        string
		stats       ffmpeg.VideoStats
		targetCodec string
		quality     Quality
		want        int
		wantErr     assert.ErrorAssertionFunc
	}{
		{
			name:    "invalid source codec",
			stats:   ffmpeg.VideoStats{VideoCodec: "invalid"},
			wantErr: assert.Error,
		},
		{
			name:        "invalid target codec",
			stats:       ffmpeg.VideoStats{VideoCodec: "h264"},
			targetCodec: "invalid",
			wantErr:     assert.Error,
		},
		{
			// low quality: 3M*0.8=2.4M minimum. target minimum is 1.5M
			name:        "low quality",
			stats:       ffmpeg.VideoStats{VideoCodec: "h264", BitRate: 4_000_000, Height: 720},
			quality:     LowQuality,
			targetCodec: "hevc",
			want:        1_500_000,
			wantErr:     assert.NoError,
		},
		{
			// target minimum is 1.5M
			name:        "high quality",
			stats:       ffmpeg.VideoStats{VideoCodec: "h264", BitRate: 4_000_000, Height: 720},
			quality:     HighQuality,
			targetCodec: "hevc",
			want:        1_500_000,
			wantErr:     assert.NoError,
		},
		{
			// minimum: 3M. 6M oversample factor: 2. target should be 1.5M*2 = 3M
			name:        "max quality",
			stats:       ffmpeg.VideoStats{VideoCodec: "h264", BitRate: 6_000_000, Height: 720},
			quality:     MaxQuality,
			targetCodec: "hevc",
			want:        3_000_000,
			wantErr:     assert.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			bitrate, err := getTargetBitRate(tt.stats, tt.targetCodec, tt.quality)
			assert.Equal(t, tt.want, bitrate)
			tt.wantErr(t, err)
		})
	}
}
