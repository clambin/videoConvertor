package pipeline

import (
	"codeberg.org/clambin/go-common/flagger"
	"github.com/spf13/viper"
)

type Configuration struct {
	flagger.Log
	Input       string  `flagger.usage:"input directory"`
	ProfileName string  `flagger.name:"profile" flagger.usage:"conversion profile"`
	Profile     Profile `flagger.skip:"true"`
	Active      bool    `flagger.usage:"start processor in active mode"`
	Remove      bool    `flagger.usage:"remove source files after successful conversion"`
	Overwrite   bool    `flagger.usage:"overwrite existing files"`
}

func GetConfigurationFromViper(v *viper.Viper) (cfg Configuration, err error) {
	cfg.Active = v.GetBool("active")
	cfg.Input = v.GetString("input")
	cfg.Format = v.GetString("log.format")
	cfg.Level = v.GetString("log.level")
	cfg.Overwrite = v.GetBool("overwrite")
	cfg.Remove = v.GetBool("remove")
	cfg.ProfileName = v.GetString("profile")
	if cfg.Profile, err = GetProfile(cfg.ProfileName); err != nil {
		return Configuration{}, err
	}
	return cfg, nil
}
