package cmd

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"time"

	"codeberg.org/clambin/go-common/charmer"
	"github.com/clambin/xcoder/ffmpeg"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	probeCmd = &cobra.Command{
		Use:          "probe",
		Short:        "Determine video properties of a media file",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			for _, arg := range args {
				stats, err := ffmpeg.Probe(arg)
				if err != nil {
					return err
				}

				if viper.GetBool("json") {
					enc := json.NewEncoder(os.Stdout)
					enc.SetIndent("", "  ")
					_ = enc.Encode(stats)
				} else {
					fmt.Printf("%s: codec:%s bitrate:%s height:%d width:%d duration: %6s\n",
						arg,
						stats.VideoCodec,
						ffmpeg.Bits(stats.BitRate).Format(1),
						stats.Height,
						stats.Width,
						(time.Duration(math.Round(stats.Duration.Seconds())) * time.Second).String(),
					)
				}
			}
			return nil
		},
	}

	probeArgs = charmer.Arguments{
		"json": {Default: false, Help: "output properties as JSON"},
	}
)

func init() {
	rootCmd.AddCommand(probeCmd)
	if err := charmer.SetPersistentFlags(probeCmd, viper.GetViper(), probeArgs); err != nil {
		panic(err)
	}
}
