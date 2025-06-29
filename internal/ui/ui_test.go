package ui

import (
	"testing"

	"github.com/clambin/xcoder/internal/pipeline"
	"github.com/gdamore/tcell/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUI(t *testing.T) {
	var list pipeline.Queue
	list.Add("foo")
	list.List()[0].SetWorkStatus(pipeline.WorkStatus{Status: pipeline.Skipped})

	cfg := pipeline.Configuration{Input: "/foo", ProfileName: "foo"}
	ui := New(&list, cfg)
	ui.refresh()

	// validate header - configuration
	const wantConfigText = ` Base directory:   /foo
 Profile:          foo
 Remove source:    off
 Overwrite target: off
`
	assert.Equal(t, wantConfigText, ui.header.configPane.GetText(true))
	// validate header - status
	const wantStatusText = ` Converting    : off
`
	assert.Equal(t, wantStatusText, ui.header.statusPane.GetText(true))

	// default view: table is unfiltered
	assert.Equal(t, 2, ui.queueViewer.GetRowCount())
	assert.Equal(t, "foo", ui.queueViewer.GetCell(1, 0).Text)

	// validate that filters work
	assert.Nil(t, ui.queueViewer.handleInput(tcell.NewEventKey(tcell.KeyRune, 's', tcell.ModNone)))
	ui.refresh()
	assert.Equal(t, 1, ui.queueViewer.GetRowCount())
	assert.Nil(t, ui.queueViewer.handleInput(tcell.NewEventKey(tcell.KeyRune, 's', tcell.ModNone)))
	ui.refresh()
	assert.Equal(t, 2, ui.queueViewer.GetRowCount())
	assert.NotNil(t, ui.queueViewer.handleInput(tcell.NewEventKey(tcell.KeyRune, 's', tcell.ModCtrl)))
	assert.Equal(t, 2, ui.queueViewer.GetRowCount())

	// verify switching between different pages
	front, _ := ui.header.shortcutsView.GetFrontPage()
	assert.Equal(t, "worklist", front)
	assert.Nil(t, ui.handleInput(tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone)))
	ui.refresh()
	front, _ = ui.header.shortcutsView.GetFrontPage()
	assert.Equal(t, "logs", front)
	assert.Nil(t, ui.handleInput(tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone)))
	ui.refresh()
	front, _ = ui.header.shortcutsView.GetFrontPage()
	assert.Equal(t, "worklist", front)

	// Request a file to be converted
	ui.queueViewer.Select(1, 0)
	list.List()[0].SetWorkStatus(pipeline.WorkStatus{Status: pipeline.Inspected})
	assert.Nil(t, ui.queueViewer.handleInput(tcell.NewEventKey(tcell.KeyEnter, rune(tcell.KeyEnter), tcell.ModNone)))
	item := list.NextToConvert()
	require.NotNil(t, item)
	assert.Equal(t, "foo", item.Source.Path)
}
