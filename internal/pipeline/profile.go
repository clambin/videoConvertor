package pipeline

import (
	"fmt"
	"maps"
	"slices"
	"strconv"
	"strings"

	"github.com/clambin/xcoder/ffmpeg"
)

var profiles = map[string]Profile{
	"hevc-low": {
		TargetCodec: "hevc",
		Rules: []Rule{
			SkipTargetCodec(),
			RejectBitrateTooLow(),
		},
		CapBitrate: true,
	},
	"hevc-medium": {
		TargetCodec: "hevc",
		Rules: []Rule{
			SkipTargetCodec(),
			RejectVideoHeightTooLow(720),
			RejectBitrateTooLow(),
		},
		CapBitrate: true,
	},
	"hevc-high": {
		TargetCodec: "hevc",
		Rules: []Rule{
			SkipTargetCodec(),
			RejectVideoHeightTooLow(1080),
			RejectBitrateTooLow(),
		},
	},
}

type Profile struct {
	TargetCodec string
	Rules       []Rule
	CapBitrate  bool
}

// GetProfile returns the profile associated with name.
func GetProfile(name string) (Profile, error) {
	if profile, ok := profiles[name]; ok {
		return profile, nil
	}
	return Profile{}, fmt.Errorf("invalid profile name: %q. supported profile names: %s", name, strings.Join(SupportedProfiles(), ", ")) //nolint:err113
}

func SupportedProfiles() []string {
	p := slices.Collect(maps.Keys(profiles))
	slices.Sort(p)
	return p
}

func (p *Profile) Inspect(item *WorkItem) (*ffmpeg.FFMPEG, error) {
	// evaluate all rules
	for _, rule := range p.Rules {
		if err := rule(p, item.Source.VideoStats); err != nil {
			return nil, err
		}
	}

	// determine filename of the output
	item.Target.Path = buildTargetFilename(item.Source, "", p.TargetCodec, "mkv")

	// determine target videoStats.  If we're asked to cap the bitrate, we ask for the minimum bitrate of the target codec.
	// otherwise, if the source bitrate is higher than the minimum, increate the target bitrate by the same factor.
	item.Target.VideoStats = item.Source.VideoStats
	item.Target.VideoStats.VideoCodec = p.TargetCodec
	var err error
	if item.Target.VideoStats.BitRate, err = getTargetBitrate(item.Source.VideoStats, item.Source.VideoStats.VideoCodec, p.TargetCodec, p.CapBitrate); err != nil {
		return nil, err
	}

	// build the transcoder
	xcoder := ffmpeg.
		Decode(item.Source.Path, decoderOptions[item.Target.VideoStats.VideoCodec]...).
		Encode(encoderArguments(item.Target.VideoStats)...).
		Muxer("matroska"). // mkv only
		NoStats().
		LogLevel("error").
		Output(item.Target.Path)

	return xcoder, nil
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type Rule func(profile *Profile, sourceStats ffmpeg.VideoStats) error

func SkipTargetCodec() Rule {
	return func(profile *Profile, sourceStats ffmpeg.VideoStats) error {
		if sourceStats.VideoCodec == profile.TargetCodec {
			return &SourceSkippedError{Reason: "source video already in target codec"}
		}
		return nil
	}
}

func RejectVideoHeightTooLow(height int) Rule {
	return func(_ *Profile, sourceStats ffmpeg.VideoStats) error {
		if sourceStats.Height < height {
			return &SourceRejectedError{Reason: "source video height is less than " + strconv.Itoa(height)}
		}
		return nil
	}
}

func RejectBitrateTooLow() Rule {
	return func(profile *Profile, sourceStats ffmpeg.VideoStats) error {
		// evaluate minimum bitrate for the source. we check both source & target codec, as target codec may need
		// a higher bitrate than the source codec (e.g. hevc -> h264).
		minimumBitrate, err := getMinimumBitRate(sourceStats, sourceStats.VideoCodec, profile.TargetCodec)
		if err != nil {
			return &SourceRejectedError{Reason: err.Error()}
		}
		if sourceStats.BitRate < minimumBitrate {
			return &SourceRejectedError{Reason: "source bitrate must be at least " + ffmpeg.Bits(minimumBitrate).Format(1)}
		}
		return nil
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type bitRate struct {
	height  int
	bitrate int
}

type bitRates []bitRate

func (b bitRates) getBitrate(height int) int {
	for i, r := range b {
		if r.height < height {
			continue
		}
		if r.height == height {
			return r.bitrate
		}
		if i == 0 {
			return r.bitrate
		}
		factor := float64(height-b[i-1].height) / float64(b[i].height-b[i-1].height)
		return b[i-1].bitrate + int(factor*(float64(b[i].bitrate-b[i-1].bitrate)))
	}
	return b[len(b)-1].bitrate
}

// https://www.yololiv.com/blog/h265-vs-h264-whats-the-difference-which-is-better/

var minimumBitrates = map[string]bitRates{
	"h264": {
		{height: 480, bitrate: 1500_000},
		{height: 720, bitrate: 3_000_000},
		{height: 1080, bitrate: 6_000_000},
		{height: 2160, bitrate: 32_000_000},
	},
	"hevc": {
		{height: 480, bitrate: 750_000},
		{height: 720, bitrate: 1_500_000},
		{height: 1080, bitrate: 3_000_000},
		{height: 2160, bitrate: 15_000_000},
	},
}

// getMinimumBitRate determines minimum bitrate for the source. we check both source & target codec, as target codec may need
// a higher bitrate than the source codec (e.g. hevc -> h264).
func getMinimumBitRate(videoStats ffmpeg.VideoStats, from string, to string) (int, error) {
	sourceMinimumBitrates, ok := minimumBitrates[from]
	if !ok {
		return 0, &UnsupportedCodecError{Codec: from}
	}
	targetMinimumBitrates, ok := minimumBitrates[to]
	if !ok {
		return 0, &UnsupportedCodecError{Codec: to}
	}
	return max(sourceMinimumBitrates.getBitrate(videoStats.Height), targetMinimumBitrates.getBitrate(videoStats.Height)), nil
}

func getTargetBitrate(videoStats ffmpeg.VideoStats, from string, to string, capBitrate bool) (int, error) {
	sourceMinimumBitrates, ok := minimumBitrates[from]
	if !ok {
		return 0, &UnsupportedCodecError{Codec: from}
	}
	targetMinimumBitrates, ok := minimumBitrates[to]
	if !ok {
		return 0, &UnsupportedCodecError{Codec: to}
	}
	bitrate := targetMinimumBitrates.getBitrate(videoStats.Height)
	if !capBitrate {
		oversampling := float64(videoStats.BitRate) / float64(sourceMinimumBitrates.getBitrate(videoStats.Height))
		bitrate = int(float64(bitrate) * oversampling)
	}
	return bitrate, nil
}
