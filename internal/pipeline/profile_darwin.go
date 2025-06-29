package pipeline

import (
	"strconv"

	"github.com/clambin/xcoder/ffmpeg"
)

var decoderOptions = map[string][]string{
	"hevc": {"-hwaccel", "videotoolbox"},
}

func encoderArguments(videoStats ffmpeg.VideoStats) []string {
	switch videoStats.VideoCodec {
	case "hevc":
		profileName := "main"
		if videoStats.BitsPerSample == 10 {
			profileName = "main10"
		}
		return []string{
			"-c:v", "hevc_videotoolbox",
			"-b:v", strconv.Itoa(videoStats.BitRate),
			"-profile:v", profileName,
			"-c:a", "copy",
			"-c:s", "copy",
		}
	default:
		return nil
	}
}
