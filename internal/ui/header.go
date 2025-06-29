package ui

import (
	"fmt"
	"strings"

	"github.com/clambin/xcoder/internal/pipeline"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type header struct {
	configPane    *configPane
	statusPane    *statusPane
	shortcutsView *shortcutsView
	*tview.Grid
}

func newHeader(list *pipeline.Queue, configuration pipeline.Configuration) *header {
	h := header{
		Grid:          tview.NewGrid(),
		configPane:    newConfigPane(configuration),
		statusPane:    newStatusPane(list),
		shortcutsView: newShortcutsView(),
	}
	h.AddItem(h.configPane, 0, 0, 1, 1, 0, 0, false)
	h.AddItem(h.statusPane, 0, 1, 1, 1, 0, 0, false)
	h.AddItem(h.shortcutsView, 0, 2, 1, 2, 0, 0, false)
	return &h
}

func (h *header) refresh() {
	h.statusPane.refresh()
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type configPane struct {
	*tview.TextView
}

func newConfigPane(cfg pipeline.Configuration) *configPane {
	p := configPane{TextView: tview.NewTextView().SetDynamicColors(true)}
	var content strings.Builder
	content.WriteString(fmt.Sprintf(" [%s]Base directory:   [%s]%s\n", labelColor, tview.Styles.SecondaryTextColor, cfg.Input))
	content.WriteString(fmt.Sprintf(" [%s]Profile:          [%s]%s\n", labelColor, tview.Styles.SecondaryTextColor, cfg.ProfileName))
	content.WriteString(fmt.Sprintf(" [%s]Remove source:    [%s]%s\n", labelColor, tview.Styles.SecondaryTextColor, onOffString[cfg.Remove]))
	content.WriteString(fmt.Sprintf(" [%s]Overwrite target: [%s]%s\n", labelColor, tview.Styles.SecondaryTextColor, onOffString[cfg.Overwrite]))
	p.SetText(content.String())
	return &p
}

var onOffString = map[bool]string{
	true:  "on",
	false: "off",
}

var onOffColor = map[bool]tcell.Color{
	true:  tcell.ColorGreen,
	false: textColor,
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type statusPane struct {
	*tview.TextView
	list *pipeline.Queue
}

func newStatusPane(list *pipeline.Queue) *statusPane {
	return &statusPane{
		TextView: tview.NewTextView().SetDynamicColors(true),
		list:     list,
	}
}

func (s *statusPane) refresh() {
	converting := s.list.Active()
	content := fmt.Sprintf(" [%s]Converting    : [%s]%s\n", labelColor, onOffColor[converting], onOffString[converting])
	s.Clear()
	s.SetText(content)
}
