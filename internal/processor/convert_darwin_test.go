package processor

import (
	"github.com/clambin/videoConvertor/ffmpeg"
	"github.com/stretchr/testify/assert"
)

var makeConvertCommandTests = []struct {
	name           string
	request        Request
	progressSocket string
	want           string
	wantErr        assert.ErrorAssertionFunc
}{
	{
		name: "hevc 8 bit",
		request: Request{
			Source:      "foo.mkv",
			Target:      "foo.hevc.mkv",
			TargetStats: ffmpeg.VideoStats{VideoCodec: "hevc", BitRate: 4_000_000, BitsPerSample: 8},
		},
		progressSocket: "socket",
		want:           "-hwaccel videotoolbox -i foo.mkv -b:v 4000000 -c:a copy -c:s copy -c:v hevc_videotoolbox -f matroska -profile:v main foo.hevc.mkv -loglevel error -nostats -progress unix://socket -y",
		wantErr:        assert.NoError,
	},
	{
		name: "hevc 10 bit",
		request: Request{
			Source:      "foo.mkv",
			Target:      "foo.hevc.mkv",
			TargetStats: ffmpeg.VideoStats{VideoCodec: "hevc", BitRate: 4_000_000, BitsPerSample: 10},
		},
		progressSocket: "socket",
		want:           "-hwaccel videotoolbox -i foo.mkv -b:v 4000000 -c:a copy -c:s copy -c:v hevc_videotoolbox -f matroska -profile:v main10 foo.hevc.mkv -loglevel error -nostats -progress unix://socket -y",
		wantErr:        assert.NoError,
	},
	{
		name: "default is 8 bit",
		request: Request{
			Source:      "foo.mkv",
			Target:      "foo.hevc.mkv",
			TargetStats: ffmpeg.VideoStats{VideoCodec: "hevc", BitRate: 4_000_000},
		},
		progressSocket: "socket",
		want:           "-hwaccel videotoolbox -i foo.mkv -b:v 4000000 -c:a copy -c:s copy -c:v hevc_videotoolbox -f matroska -profile:v main foo.hevc.mkv -loglevel error -nostats -progress unix://socket -y",
		wantErr:        assert.NoError,
	},
	{
		name: "no progress socket",
		request: Request{
			Source:      "foo.mkv",
			Target:      "foo.hevc.mkv",
			TargetStats: ffmpeg.VideoStats{VideoCodec: "hevc", BitRate: 4_000_000, BitsPerSample: 8},
		},
		want:    "-hwaccel videotoolbox -i foo.mkv -b:v 4000000 -c:a copy -c:s copy -c:v hevc_videotoolbox -f matroska -profile:v main foo.hevc.mkv -loglevel error -nostats -y",
		wantErr: assert.NoError,
	},
	{
		name: "only support for hevc",
		request: Request{
			Source:      "foo.mkv",
			Target:      "foo.hevc.mkv",
			TargetStats: ffmpeg.VideoStats{VideoCodec: "h264", BitRate: 4_000_000, BitsPerSample: 8},
		},
		wantErr: assert.Error,
	},
}
