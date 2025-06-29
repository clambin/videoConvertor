package ui

import (
	"log/slog"
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/stretchr/testify/assert"
)

func Test_logViewer(t *testing.T) {
	v := newLogViewer()

	l := slog.New(slog.NewTextHandler(v, &slog.HandlerOptions{}))
	l.Info("hello world")

	assert.Contains(t, v.GetText(true), "hello world")

	assert.False(t, v.wrap)
	for _, want := range []bool{true, false} {
		v.handleInput(tcell.NewEventKey(tcell.KeyRune, 'w', tcell.ModNone))
		assert.Equal(t, want, v.wrap)
	}

	for _, key := range []*tcell.EventKey{
		tcell.NewEventKey(tcell.KeyRune, 'w', tcell.ModCtrl),
		tcell.NewEventKey(tcell.KeyRune, 'e', tcell.ModNone),
	} {
		v.handleInput(key)
		assert.False(t, v.wrap)
	}
}
