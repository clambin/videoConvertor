package profile

import (
	"fmt"
	"github.com/clambin/videoConvertor/internal/ffmpeg"
)

var profiles = map[string]Profile{
	"hevc-low": {
		Codec:              "hevc",
		ConstantRateFactor: 28,
		Rules: Rules{
			SkipCodec("hevc"),
			MinimumBitrate(0.8),
		},
	},
	"hevc-medium": {
		Codec:              "hevc",
		ConstantRateFactor: 18,
		Rules: Rules{
			SkipCodec("hevc"),
			MinimumBitrate(1),
		},
	},
	"hevc-high": {
		Codec:              "hevc",
		ConstantRateFactor: 10,
		Rules: Rules{
			SkipCodec("hevc"),
			MinimumHeight(720),
			MinimumBitrate(1),
		},
	},
}

// A Profile serves two purposes. Firstly, it evaluates whether a source video file meets the requirements to be converted.
// Secondly, it determines the video parameters of the output video file.
type Profile struct {
	Codec              string
	Rules              Rules
	ConstantRateFactor int
}

// GetProfile returns the profile associated with name.
func GetProfile(name string) (Profile, error) {
	if profile, ok := profiles[name]; ok {
		return profile, nil
	}
	return Profile{}, fmt.Errorf("invalid profile name: %s", name)
}

// Evaluate verifies that the source's videoStats meet the profile's requirements. Otherwise it returns the first non-compliance.
func (p Profile) Evaluate(sourceVideoStats ffmpeg.VideoStats) error {
	return p.Rules.ShouldConvert(sourceVideoStats)
}
