package ffmpeg_test

import (
	"strings"
	"testing"

	"github.com/clambin/videoConvertor/ffmpeg"
	"github.com/stretchr/testify/assert"
)

func TestFFMPEG_Build(t *testing.T) {
	cmd := ffmpeg.Input("foo.mkv", ffmpeg.Args{"hwaccel": "hevc_videotoolbox"}).
		Output("foo.hevc", ffmpeg.Args{
			"c:v":       "hevc_videotoolbox",
			"profile:v": "main",
			"crf":       "10",
			"c:a":       "copy",
			"c:s":       "copy",
			"f":         "matroska",
		}).
		LogLevel("error").
		NoStats().
		OverWriteTarget().
		ProgressSocket("socket").
		AddGlobalArguments(ffmpeg.Args{"foo": "bar"})

	want := `-hwaccel hevc_videotoolbox -i foo.mkv -c:a copy -c:s copy -c:v hevc_videotoolbox -crf 10 -f matroska -profile:v main foo.hevc -foo bar -loglevel error -nostats -progress unix://socket -y`
	assert.Equal(t, want, strings.Join(cmd.Build(t.Context()).Args[1:], " "))
}
