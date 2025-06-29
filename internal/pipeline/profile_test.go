package pipeline

import (
	"testing"

	"github.com/clambin/xcoder/ffmpeg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProfiles(t *testing.T) {
	valid1080 := &WorkItem{Source: MediaFile{Path: "foo.mkv", VideoStats: ffmpeg.VideoStats{VideoCodec: "h264", BitRate: 8_000_000, Height: 1080}}}
	tooLow1080 := &WorkItem{Source: MediaFile{Path: "foo.mkv", VideoStats: ffmpeg.VideoStats{VideoCodec: "h264", BitRate: 1_000_000, Height: 1080}}}
	valid720 := &WorkItem{Source: MediaFile{Path: "foo.mkv", VideoStats: ffmpeg.VideoStats{VideoCodec: "h264", BitRate: 8_000_000, Height: 720}}}
	tooLow720 := &WorkItem{Source: MediaFile{Path: "foo.mkv", VideoStats: ffmpeg.VideoStats{VideoCodec: "h264", BitRate: 750_000, Height: 720}}}
	valid360 := &WorkItem{Source: MediaFile{Path: "foo.mkv", VideoStats: ffmpeg.VideoStats{VideoCodec: "h264", BitRate: 8_000_000, Height: 360}}}
	tooLow360 := &WorkItem{Source: MediaFile{Path: "foo.mkv", VideoStats: ffmpeg.VideoStats{VideoCodec: "h264", BitRate: 100_000, Height: 360}}}

	tests := map[string][]struct {
		name    string
		source  *WorkItem
		wantErr assert.ErrorAssertionFunc
	}{
		"hevc-high": {
			{"valid 1080", valid1080, assert.NoError},
			{"1080 bitrate too low", tooLow1080, assert.Error},
			{"valid 720", valid720, assert.Error},
			{"720 bitrate too low", tooLow720, assert.Error},
			{"valid 360", valid360, assert.Error},
			{"360 bitrate too low", tooLow360, assert.Error},
		},
		"hevc-medium": {
			{"valid 1080", valid1080, assert.NoError},
			{"1080 bitrate too low", tooLow1080, assert.Error},
			{"valid 720", valid720, assert.NoError},
			{"720 bitrate too low", tooLow720, assert.Error},
			{"valid 360", valid360, assert.Error},
			{"360 bitrate too low", tooLow360, assert.Error},
		},
		"hevc-low": {
			{"valid 1080", valid1080, assert.NoError},
			{"1080 bitrate too low", tooLow1080, assert.Error},
			{"valid 720", valid720, assert.NoError},
			{"720 bitrate too low", tooLow720, assert.Error},
			{"valid 360", valid360, assert.NoError},
			{"360 bitrate too low", tooLow360, assert.Error},
		},
	}
	for name, cases := range tests {
		p, err := GetProfile(name)
		require.NoError(t, err)
		t.Run(name, func(t *testing.T) {
			for _, testCase := range cases {
				t.Run(testCase.name, func(t *testing.T) {
					_, err = p.Inspect(testCase.source)
					testCase.wantErr(t, err)
				})
			}
		})
	}
}

func TestGetProfile(t *testing.T) {
	tests := []struct {
		name      string
		wantErr   assert.ErrorAssertionFunc
		wantCodec string
	}{
		{"hevc-high", assert.NoError, "hevc"},
		{"hevc-medium", assert.NoError, "hevc"},
		{"hevc-low", assert.NoError, "hevc"},
		{"invalid", assert.Error, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			profile, err := GetProfile(tt.name)
			tt.wantErr(t, err)
			assert.Equal(t, tt.wantCodec, profile.TargetCodec)
		})
	}
}

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
