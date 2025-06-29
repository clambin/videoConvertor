package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/clambin/xcoder/ffmpeg"
	"github.com/spf13/cobra"
)

var (
	verifyCmd = &cobra.Command{
		Use:   "verify",
		Short: "Verify media files",
		RunE: func(cmd *cobra.Command, args []string) error {
			for _, arg := range args {
				verify(cmd.Context(), arg)
			}
			return nil
		},
	}
)

func init() {
	rootCmd.AddCommand(verifyCmd)
}

func verify(ctx context.Context, path string) {
	stats, err := ffmpeg.Probe(path)
	if err != nil {
		fmt.Printf("\r%s FAIL: %v\n", path, err)
		return
	}

	tmpDir, err := os.MkdirTemp(os.TempDir(), "xcoder")
	if err != nil {
		panic(err)
	}
	tempSocketPath := filepath.Join(tmpDir, "ffmpeg-verify.sock")

	f := ffmpeg.Decode(path, "-hwaccel", "videotoolbox").Muxer("null").LogLevel("error").NoStats().Progress(func(p ffmpeg.Progress) {
		progress := p.Converted.Seconds() / stats.Duration.Seconds()
		fmt.Printf("\r%s ... %.1f%%", path, 100*progress)
	}, tempSocketPath)

	err = f.Run(ctx, slog.Default())
	if err == nil {
		fmt.Println(" PASS")
	} else {
		fmt.Println(" FAIL: ", err.Error())
	}
}
