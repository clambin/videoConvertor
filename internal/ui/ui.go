package ui

import (
	"context"
	"time"

	"github.com/clambin/videoConvertor/internal/configuration"
	"github.com/clambin/videoConvertor/internal/pipeline"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type UI struct {
	Root        *tview.Grid
	pages       *tview.Pages
	header      *header
	queueViewer *queueViewer
	LogViewer   *LogViewer
}

type Application interface {
	QueueUpdateDraw(func()) *tview.Application
}

func New(list *pipeline.Queue, cfg configuration.Configuration) *UI {
	h := newHeader(list, cfg)
	b := tview.NewPages()

	wlv := newQueueViewer(list)
	b.AddPage("worklist", wlv, true, true)
	h.shortcutsView.addPage("worklist", workListShortCuts, true)
	lv := newLogViewer()
	b.AddPage("logs", lv, true, false)
	h.shortcutsView.addPage("logs", logViewerShortCuts, false)

	root := tview.NewGrid()
	root.AddItem(h, 0, 0, 1, 1, 0, 0, false)
	root.AddItem(b, 1, 0, 3, 1, 0, 0, true)

	u := UI{
		Root:        root,
		pages:       b,
		header:      h,
		queueViewer: wlv,
		LogViewer:   lv,
	}

	u.Root.SetInputCapture(u.handleInput)
	return &u
}

func (u *UI) Run(ctx context.Context, app Application, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			app.QueueUpdateDraw(func() {
				u.refresh()
			})
		}
	}
}

func (u *UI) refresh() {
	u.header.refresh()
	u.queueViewer.refresh()
}

func (u *UI) handleInput(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyRune:
		switch event.Rune() {
		case 'l':
			page, _ := u.pages.GetFrontPage()
			switch page {
			case "worklist":
				u.pages.HidePage("worklist")
				u.pages.ShowPage("logs")
				u.header.shortcutsView.HidePage("worklist")
				u.header.shortcutsView.ShowPage("logs")
			case "logs":
				u.pages.HidePage("logs")
				u.pages.ShowPage("worklist")
				u.header.shortcutsView.HidePage("logs")
				u.header.shortcutsView.ShowPage("worklist")
			}
			return nil
		default:
			return event
		}
	default:
		return event
	}
}
