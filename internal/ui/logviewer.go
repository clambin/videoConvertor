package ui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var logViewerShortCuts = shortcutsPage{
	{shortcut{key: "l", description: "close logs"}, shortcut{key: "w", description: "wrap lines"}},
}

type LogViewer struct {
	*tview.TextView
	wrap bool
}

func newLogViewer() *LogViewer {
	v := LogViewer{TextView: tview.NewTextView()}
	v.
		SetBorder(true).
		SetTitle(" logs ").
		SetTitleAlign(tview.AlignCenter).
		SetBorderPadding(0, 0, 1, 1).
		SetInputCapture(v.handleInput)
	v.
		SetScrollable(true).
		ScrollToEnd().
		SetWrap(v.wrap)
	return &v
}

func (v *LogViewer) ToggleWrap() {
	v.wrap = !v.wrap
	v.SetWrap(v.wrap)
}

func (v *LogViewer) handleInput(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyRune:
		switch event.Rune() {
		case 'w':
			if event.Modifiers() == tcell.ModNone {
				v.ToggleWrap()
				return nil
			}
		default:
		}
	default:
		return event
	}
	return event
}
