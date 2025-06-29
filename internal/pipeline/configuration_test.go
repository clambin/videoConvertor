package pipeline

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetConfigurationFromViper(t *testing.T) {
	v := viper.New()
	v.Set("profile", "invalid")

	_, err := GetConfigurationFromViper(v)
	require.Error(t, err)

	v.Set("profile", "hevc-high")
	cfg, err := GetConfigurationFromViper(v)
	require.NoError(t, err)
	assert.Equal(t, "hevc", cfg.Profile.TargetCodec)
}
