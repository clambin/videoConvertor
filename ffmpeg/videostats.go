package ffmpeg

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"strconv"
	"strings"
	"time"
)

//nolint:tagliatelle
type VideoStats struct {
	VideoCodec    string        `json:"video_codec"`
	Duration      time.Duration `json:"duration"`
	BitRate       int           `json:"bit_rate"`
	BitsPerSample int           `json:"bits_per_sample"`
	Height        int           `json:"height"`
	Width         int           `json:"width"`
}

func parseVideoStats(r io.Reader) (VideoStats, error) {
	//nolint:tagliatelle
	var stats struct {
		Format struct {
			Filename string `json:"filename"`
			Duration string `json:"duration"`
			BitRate  string `json:"bit_rate"`
		} `json:"format"`
		Streams []struct {
			CodecName        string `json:"codec_name,omitempty"`
			CodecType        string `json:"codec_type"`
			BitsPerRawSample string `json:"bits_per_raw_sample,omitempty"`
			Height           int    `json:"height,omitempty"`
			Width            int    `json:"width,omitempty"`
		} `json:"streams"`
	}

	if err := json.NewDecoder(r).Decode(&stats); err != nil {
		return VideoStats{}, fmt.Errorf("json: %w", err)
	}

	var videoStats VideoStats
	duration, err := strconv.ParseFloat(stats.Format.Duration, 64)
	if err != nil {
		return VideoStats{}, fmt.Errorf("invalid duration: %w", err)
	}
	videoStats.Duration = time.Duration(duration*1000) * time.Millisecond
	bitrate, err := strconv.Atoi(stats.Format.BitRate)
	if err != nil {
		return VideoStats{}, fmt.Errorf("invalid bit_rate: %w", err)
	}
	videoStats.BitRate = bitrate

	for _, stream := range stats.Streams {
		if stream.CodecType == "video" {
			videoStats.VideoCodec = stream.CodecName
			videoStats.Height = stream.Height
			videoStats.Width = stream.Width
			switch stream.BitsPerRawSample {
			case "", "8":
				videoStats.BitsPerSample = 8
			case "10":
				videoStats.BitsPerSample = 10
			default:
				return VideoStats{}, &InvalidMediaError{Reason: "invalid bits_per_raw_sample: " + stream.BitsPerRawSample}
			}
		}
	}

	if videoStats.VideoCodec == "" {
		return VideoStats{}, &InvalidMediaError{Reason: "no video stream found"}
	}

	return videoStats, nil
}

func (s VideoStats) String() string {
	if s.VideoCodec == "" {
		return ""
	}
	output := make([]string, 1, 3)
	output[0] = s.VideoCodec
	if s.Height > 0 {
		output = append(output, strconv.Itoa(s.Height))
	}
	if s.BitRate > 0 {
		output = append(output, Bits(s.BitRate).Format(2))
	}
	return strings.Join(output, "/")
}

var _ slog.LogValuer = VideoStats{}

func (s VideoStats) LogValue() slog.Value {
	values := make([]slog.Attr, 0, 4)
	if s.VideoCodec != "" {
		values = append(values, slog.String("codec", s.VideoCodec))
	}
	if s.Height > 0 || s.Width > 0 {
		values = append(values, slog.Int("width", s.Width), slog.Int("height", s.Height))
	}
	if s.BitRate > 0 {
		values = append(values, slog.String("bitrate", Bits(s.BitRate).Format(1)))
	}
	return slog.GroupValue(values...)
}

type Bits int

func (b Bits) Format(decimals int) string {
	floatBits := float64(b)
	unit := "b"
	if floatBits > 1000 {
		floatBits /= 1000
		unit = "kb"
	}
	if floatBits > 1000 {
		floatBits /= 1000
		unit = "mb"
	}
	return strconv.FormatFloat(floatBits, 'f', decimals, 64) + " " + unit + "ps"
}

type InvalidMediaError struct {
	Reason string
}

func (e InvalidMediaError) Error() string {
	return e.Reason
}
