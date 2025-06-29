package ffmpeg

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_parseVideoStats(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    VideoStats
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "valid",
			input: `{
    "streams": [
        { "codec_name": "hevc", "codec_type": "video", "bits_per_raw_sample": "10", "height": 1080, "width": 1920 },
        { "codec_type": "audio" },
        { "codec_type": "subtitle" }
    ],
    "format": { "filename": "foo.hevc.mkv", "duration": "1800.000", "bit_rate": "5000000" }
}
`,
			want:    VideoStats{Duration: 30 * time.Minute, VideoCodec: "hevc", BitRate: 5_000_000, BitsPerSample: 10, Height: 1080, Width: 1920},
			wantErr: assert.NoError,
		},
		{
			name:    "bitsPerSample defaults to 8",
			input:   `{"format": { "duration": "1800.000", "bit_rate": "5000000" }, "streams": [ { "codec_name": "hevc", "codec_type": "video", "height": 1080, "width": 1920 } ]}`,
			want:    VideoStats{Duration: 30 * time.Minute, VideoCodec: "hevc", BitRate: 5_000_000, BitsPerSample: 8, Height: 1080, Width: 1920},
			wantErr: assert.NoError,
		},
		{
			name:    "empty",
			wantErr: assert.Error,
		},
		{
			name:    "missing duration",
			input:   `{"format": {  }}`,
			wantErr: assert.Error,
		},
		{
			name:    "invalid duration",
			input:   `{"format": { "duration": "foobar" }}`,
			wantErr: assert.Error,
		},
		{
			name:    "missing bitrate",
			input:   `{"format": { "duration": "1800.00" }}`,
			wantErr: assert.Error,
		},
		{
			name:    "invalid bitrate",
			input:   `{"format": { "duration": "1800.000", "bit_rate": "foobar" }}`,
			wantErr: assert.Error,
		},
		{
			name:    "missing streams",
			input:   `{"format": { "duration": "1800.000", "bit_rate": "5000000" }}`,
			wantErr: assert.Error,
		},
		{
			name:    "missing video stream",
			input:   `{"format": { "duration": "1800.000", "bit_rate": "5000000" }, "streams": []}`,
			wantErr: assert.Error,
		},
		{
			name:    "invalid bitsPerSample defaults to 8",
			input:   `{"format": { "duration": "1800.000", "bit_rate": "5000000" }, "streams": [ { "codec_name": "hevc", "codec_type": "video", "bits_per_raw_sample": "foobar", "height": 1080, "width": 1920 } ]}`,
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := parseVideoStats(strings.NewReader(tt.input))
			assert.Equal(t, tt.want, got)
			tt.wantErr(t, err)
		})
	}
}

func TestVideoStats_String(t *testing.T) {
	stats := VideoStats{
		Duration:      30 * time.Minute,
		VideoCodec:    "hevc",
		BitRate:       5_000_000,
		BitsPerSample: 10,
		Height:        1080,
	}
	const want = `hevc/1080/5.00 mbps`
	assert.Equal(t, want, stats.String())
	stats.VideoCodec = ""
	assert.Empty(t, stats.String())
}

func TestVideoStats_LogValue(t *testing.T) {
	tests := []struct {
		name       string
		videoStats VideoStats
		want       string
	}{
		{
			name:       "blank",
			videoStats: VideoStats{},
			want:       "[]",
		},
		{
			name:       "codec only",
			videoStats: VideoStats{VideoCodec: "hevc"},
			want:       "[codec=hevc]",
		},
		{
			name:       "height only",
			videoStats: VideoStats{Height: 720},
			want:       "[width=0 height=720]",
		},
		{
			name:       "width only",
			videoStats: VideoStats{Width: 1920},
			want:       "[width=1920 height=0]",
		},
		{
			name:       "bitrate only",
			videoStats: VideoStats{BitRate: 3_000_000},
			want:       "[bitrate=3.0 mbps]",
		},
		{
			name:       "complete",
			videoStats: VideoStats{VideoCodec: "hevc", Width: 1920, Height: 1080, BitRate: 3_000_000},
			want:       "[codec=hevc width=1920 height=1080 bitrate=3.0 mbps]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, tt.videoStats.LogValue().String())
		})
	}
}

func BenchmarkParse(b *testing.B) {
	// Current:
	// BenchmarkParse-10    	  623712	      1919 ns/op	    1664 B/op	      24 allocs/op
	const input = `{
		"streams": [
	{ "codec_name": "hevc", "codec_type": "video", "bits_per_raw_sample": "10", "height": 1080, "width": 1920 },
	{ "codec_type": "audio" },
	{ "codec_type": "subtitle" }
	],
	"format": { "filename": "foo.hevc.mkv", "duration": "1800.000", "bit_rate": "5000000" }
}`

	b.ReportAllocs()
	for b.Loop() {
		if _, err := parseVideoStats(strings.NewReader(input)); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkVideoStatsString(b *testing.B) {
	// Current:
	// BenchmarkVideoStatsString-10    	10126482	       107.1 ns/op	      64 B/op	       4 allocs/op
	videoStats := VideoStats{
		Width:         1920,
		VideoCodec:    "hevc",
		Duration:      time.Hour,
		BitRate:       4_000_000,
		BitsPerSample: 8,
		Height:        1080,
	}
	b.ReportAllocs()
	for b.Loop() {
		if stats := videoStats.String(); stats == "" {
			b.Fatal("empty stats")
		}
	}
}
