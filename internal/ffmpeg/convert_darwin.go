package ffmpeg

import (
	ffmpeg "github.com/u2takey/ffmpeg-go"
)

// videoCodecs translates generic codec names into the OS-specific codec names
var videoCodecs = map[string]string{
	"hevc": "hevc_videotoolbox",
}

// inputArguments are added before the input chain.
// Use MacOS videotoolbox to decode video streams.
var inputArguments = ffmpeg.KwArgs{
	"hwaccel": "videotoolbox",
}
