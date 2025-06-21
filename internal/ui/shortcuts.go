package ui

import (
	"fmt"
	"github.com/rivo/tview"
	"strings"
)

type shortcut struct {
	key         string
	description string
}
type shortcuts []shortcut
type shortcutsPage []shortcuts

type shortcutsView struct {
	*tview.Pages
}

func newShortcutsView() *shortcutsView {
	v := shortcutsView{
		Pages: tview.NewPages(),
	}
	return &v
}

func (v *shortcutsView) addPage(name string, keys shortcutsPage, visible bool) {
	v.AddPage(name, v.buildGrid(keys), true, visible)
}

func (v *shortcutsView) buildGrid(keys shortcutsPage) *tview.Grid {
	cols := make([]strings.Builder, len(keys))
	for i, col := range keys {
		for _, entry := range col {
			cols[i].WriteString(fmt.Sprintf("[%s]<%s> [%s]%s\n", shortcutColor, entry.key, tview.Styles.TertiaryTextColor, entry.description))
		}
	}

	g := tview.NewGrid()
	for i, page := range cols {
		c := tview.NewTextView().SetDynamicColors(true).SetText(page.String())
		g.AddItem(c, 0, i, 1, 1, 0, 0, false)
	}
	return g
}
