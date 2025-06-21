package configuration

import (
	"flag"
	"fmt"
	"os"

	"codeberg.org/clambin/go-common/flagger"
	"github.com/clambin/videoConvertor/internal/profile"
)

type Configuration struct {
	flagger.Log
	Input       string          `flagger.usage:"input directory"`
	ProfileName string          `flagger.name:"profile" flagger.usage:"conversion profile"`
	Profile     profile.Profile `flagger.skip:"true"`
	Active      bool            `flagger.usage:"start processor in active mode"`
	Remove      bool            `flagger.usage:"remove source files after successful conversion"`
	Overwrite   bool            `flagger.usage:"overwrite existing files"`
}

func GetConfiguration() (Configuration, error) {
	return getConfigurationWithFlagSet(flag.CommandLine, os.Args[1:]...)
}

func getConfigurationWithFlagSet(f *flag.FlagSet, args ...string) (Configuration, error) {
	cfg := Configuration{
		Input:       "/media",
		ProfileName: "hevc-max",
	}
	flagger.SetFlags(f, &cfg)
	if err := f.Parse(args); err != nil {
		return Configuration{}, err
	}

	var err error
	if cfg.Profile, err = profile.GetProfile(cfg.ProfileName); err != nil {
		return Configuration{}, fmt.Errorf("invalid profile %q: %w", cfg.ProfileName, err)
	}

	return cfg, nil
}
