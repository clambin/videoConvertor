package pipeline

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSourceRejectedError(t *testing.T) {
	err := &SourceRejectedError{Reason: "test"}
	assert.Equal(t, "test", err.Error())
	assert.ErrorIs(t, err, &SourceRejectedError{Reason: err.Error()})
}

func TestSourceSkippedError(t *testing.T) {
	err := &SourceSkippedError{Reason: "test"}
	assert.Equal(t, "test", err.Error())
	assert.ErrorIs(t, err, &SourceSkippedError{Reason: err.Error()})
}

func TestUnsupportedCodecError(t *testing.T) {
	err := &UnsupportedCodecError{Codec: "test"}
	assert.Equal(t, "unsupported codec: test", err.Error())
	assert.ErrorIs(t, err, &UnsupportedCodecError{Codec: "test"})
}
