package configuration

import (
	"flag"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetConfiguration(t *testing.T) {
	f := flag.NewFlagSet("", flag.ContinueOnError)
	cfg, err := getConfigurationWithFlagSet(f)
	assert.NoError(t, err)
	assert.Equal(t, "hevc-max", cfg.ProfileName)

	f = flag.NewFlagSet("", flag.ContinueOnError)
	_, err = getConfigurationWithFlagSet(f, "-profile", "invalid")
	assert.Error(t, err)
}
