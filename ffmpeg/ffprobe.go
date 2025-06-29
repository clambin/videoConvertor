package ffmpeg

import (
	"bytes"
	"fmt"
	"os/exec"
)

func Probe(path string) (VideoStats, error) {
	cmd := exec.Command("ffprobe",
		"-show_format",
		"-show_streams",
		"-loglevel", "error",
		"-output_format", "json",
		path,
	)
	var stdOut, stdErr bytes.Buffer
	cmd.Stdout = &stdOut
	cmd.Stderr = &stdErr
	if err := cmd.Run(); err != nil {
		return VideoStats{}, fmt.Errorf("[%s] %w", stdErr.String(), err)
	}
	return parseVideoStats(&stdOut)
}
