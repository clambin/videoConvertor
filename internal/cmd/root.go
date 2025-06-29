package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	rootCmd = cobra.Command{
		Use:   "xcoder",
		Short: "Transcode media files",
	}

	configFilename string
)

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	buildInfo, ok := debug.ReadBuildInfo()
	if !ok {
		panic("could not read build info")
	}
	rootCmd.Version = buildInfo.Main.Version

	cobra.OnInitialize(initConfig)
	rootCmd.Flags().StringVar(&configFilename, "config", "", "Configuration file")
}

func initConfig() {
	if configFilename != "" {
		viper.SetConfigFile(configFilename)
	} else {
		viper.AddConfigPath(mustConfigDir())
		viper.SetConfigName("config")
	}
	if err := viper.ReadInConfig(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "failed to read config file: "+err.Error())
	}
}

func mustConfigDir() string {
	cfgDir, err := os.UserConfigDir()
	if err != nil {
		panic("failed to get user config dir: " + err.Error())
	}
	cfgDir = filepath.Join(cfgDir, "com.github.clambin.xcoder")
	if err = os.MkdirAll(cfgDir, os.ModePerm); err != nil {
		panic("failed to create config dir: " + err.Error())
	}
	return cfgDir
}
