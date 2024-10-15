package ffmpeg

// videoCodecs translates generic codec names into the OS-specific codec names
var videoCodecs = map[string]string{
	"hevc": "libx265",
}

// inputArguments are added before the input chain.
var inputArguments = cmd.Args{}
