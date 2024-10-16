package profile

import (
	"fmt"
	"github.com/clambin/videoConvertor/internal/ffmpeg"
)

type Quality int

const (
	LowQuality Quality = iota
	HighQuality
	MaxQuality
)

var profiles = map[string]Profile{
	"hevc-low": {
		Codec:   "hevc",
		Quality: LowQuality,
		Rules: Rules{
			SkipCodec("hevc"),
			MinimumBitrate(LowQuality),
		},
	},
	"hevc-high": {
		Codec:   "hevc",
		Quality: HighQuality,
		Rules: Rules{
			SkipCodec("hevc"),
			MinimumBitrate(HighQuality),
		},
	},
	"hevc-max": {
		Codec:   "hevc",
		Quality: MaxQuality,
		Rules: Rules{
			SkipCodec("hevc"),
			MinimumHeight(720),
			MinimumBitrate(MaxQuality),
		},
	},
}

// A Profile serves two purposes. Firstly, it evaluates whether a source video file meets the requirements to be converted.
// Secondly, it determines the video parameters of the output video file.
type Profile struct {
	Codec   string
	Rules   Rules
	Quality Quality
	Bitrate int
}

// GetProfile returns the profile associated with name.
func GetProfile(name string) (Profile, error) {
	if profile, ok := profiles[name]; ok {
		return profile, nil
	}
	return Profile{}, fmt.Errorf("invalid profile name: %s", name)
}

// Evaluate verifies that the source's videoStats meet the profile's requirements and returns the target videoStats, in line with the profile's parameters.
// If the source's videoStats do not meet the profile's requirements, error indicates the reason.
// Otherwise, it returns the first error encountered.
func (p Profile) Evaluate(sourceVideoStats ffmpeg.VideoStats) (ffmpeg.VideoStats, error) {
	if err := p.Rules.ShouldConvert(sourceVideoStats); err != nil {
		return ffmpeg.VideoStats{}, err
	}
	var stats ffmpeg.VideoStats
	rate, err := p.getTargetBitRate(sourceVideoStats)
	if err == nil {
		stats = ffmpeg.VideoStats{
			VideoCodec:    p.Codec,
			BitRate:       rate,
			BitsPerSample: sourceVideoStats.BitsPerSample,
			Height:        sourceVideoStats.Height,
		}
	}
	return stats, err
}

func (p Profile) getTargetBitRate(videoStats ffmpeg.VideoStats) (int, error) {
	return getTargetBitRate(videoStats, p.Codec, p.Quality)
}

////////////////////////////////////////////////////////////////////////////////////////////////////////

type Rule func(stats ffmpeg.VideoStats) error

type Rules []Rule

func (r Rules) ShouldConvert(stats ffmpeg.VideoStats) error {
	for _, rule := range r {
		if err := rule(stats); err != nil {
			return err
		}
	}
	return nil
}

// SkipCodec rejects any video with the specified video codec
func SkipCodec(codec string) Rule {
	return func(stats ffmpeg.VideoStats) error {
		if stats.VideoCodec != codec {
			return nil
		}
		return ErrSourceInTargetCodec
	}
}

// MinimumBitrate rejects any source video with a bitrate lower than the codec's recommended bitrate for the provided Quality
func MinimumBitrate(quality Quality) Rule {
	return func(stats ffmpeg.VideoStats) error {
		minBitRate, err := getMinimumBitRate(stats, quality)
		if err != nil {
			return ErrSourceRejected{Reason: err.Error()}
		}
		if sourceBitRate := stats.BitRate; sourceBitRate < minBitRate {
			return ErrSourceRejected{Reason: "bitrate too low"}
		}
		return nil
	}
}

// MinimumHeight rejects any video with a height lower than the specified height
func MinimumHeight(minHeight int) Rule {
	return func(stats ffmpeg.VideoStats) error {
		if stats.Height < minHeight {
			return ErrSourceRejected{Reason: "height too low"}
		}
		return nil
	}
}
