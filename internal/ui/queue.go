package ui

import (
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"codeberg.org/clambin/go-common/set"
	"github.com/clambin/xcoder/internal/pipeline"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var workListShortCuts = shortcutsPage{
	{
		shortcut{key: "p", description: "enable/disable processing"},
		shortcut{key: "enter", description: "convert selected file"},
		shortcut{key: "l", description: "show logs"},
	},
	{
		shortcut{key: "s", description: "show/hide skipped files"},
		shortcut{key: "r", description: "show/hide rejected files"},
		shortcut{key: "c", description: "show/hide converted files"},
		shortcut{key: "f", description: "show/hide full path name"},
	},
}

type queueViewer struct {
	*Table
	queue *pipeline.Queue
	dataSource
}

func newQueueViewer(list *pipeline.Queue) *queueViewer {
	source := &workItems{
		list:    list,
		filters: filters{statuses: set.New[pipeline.Status]()},
	}
	v := queueViewer{
		Table:      NewTable(source),
		queue:      list,
		dataSource: source,
	}
	v.SetInputCapture(v.handleInput)
	return &v
}

func (v *queueViewer) refresh() {
	v.Table.Update()
}

func (v *queueViewer) handleInput(event *tcell.EventKey) *tcell.EventKey {
	if v.HandleInput(event) == nil {
		return nil
	}
	switch event.Key() {
	case tcell.KeyRune:
		switch event.Rune() {
		case 'p':
			v.queue.ToggleActive()
			return nil
		default:
			return event
		}
	case tcell.KeyEnter:
		row, _ := v.GetSelection()
		item := v.GetCell(row, 0).GetReference().(*pipeline.WorkItem)
		if status := item.WorkStatus().Status; status == pipeline.Inspected || status == pipeline.Failed {
			v.queue.Queue(item)
		}
		return nil
	default:
		return event
	}
}

/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type filters struct {
	statuses set.Set[pipeline.Status]
	changed  bool
	lock     sync.RWMutex
}

func (f *filters) toggle(statuses ...pipeline.Status) {
	f.lock.Lock()
	defer f.lock.Unlock()
	for _, status := range statuses {
		if f.statuses.Contains(status) {
			f.statuses.Remove(status)
		} else {
			f.statuses.Add(status)
		}
		f.changed = true
	}
}

func (f *filters) on(status pipeline.Status) bool {
	f.lock.RLock()
	defer f.lock.RUnlock()
	return f.statuses.Contains(status)
}

func (f *filters) list() []pipeline.Status {
	f.lock.RLock()
	defer f.lock.RUnlock()
	return f.statuses.ListOrdered()
}

func (f *filters) format() string {
	filtered := f.list()
	if len(filtered) == 0 {
		return ""
	}
	fs := make([]string, len(filtered))
	for i, e := range filtered {
		fs[i] = e.String()
	}
	slices.Sort(fs)
	return strings.Join(fs, ", ")
}

func (f *filters) updated() bool {
	f.lock.Lock()
	defer f.lock.Unlock()
	changed := f.changed
	f.changed = false
	return changed
}

/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func colorStatus(status pipeline.Status) tcell.Color {
	if statusColor, ok := tableColorStatus[status]; ok {
		return statusColor
	}
	return tview.Styles.PrimaryTextColor
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////

var _ dataSource = &workItems{}

type workItems struct {
	list     *pipeline.Queue
	filters  filters
	fullName atomic.Bool
}

func (w *workItems) Header() []*tview.TableCell {
	return []*tview.TableCell{
		getTableCell("SOURCE", tview.Styles.SecondaryTextColor, tview.Styles.PrimitiveBackgroundColor, tview.AlignLeft).SetSelectable(false).SetExpansion(1),
		getTableCell(padString("SOURCE STATS", 21), tview.Styles.SecondaryTextColor, tview.Styles.PrimitiveBackgroundColor, tview.AlignLeft).SetSelectable(false),
		getTableCell(padString("TARGET STATS", 21), tview.Styles.SecondaryTextColor, tview.Styles.PrimitiveBackgroundColor, tview.AlignLeft).SetSelectable(false),
		getTableCell(padString("STATUS", 10), tview.Styles.SecondaryTextColor, tview.Styles.PrimitiveBackgroundColor, tview.AlignLeft).SetSelectable(false),
		getTableCell("COMPLETED", tview.Styles.SecondaryTextColor, tview.Styles.PrimitiveBackgroundColor, tview.AlignLeft).SetSelectable(false),
		getTableCell("REMAINING", tview.Styles.SecondaryTextColor, tview.Styles.PrimitiveBackgroundColor, tview.AlignLeft).SetSelectable(false),
		getTableCell(padString("ERROR", 30), tview.Styles.SecondaryTextColor, tview.Styles.PrimitiveBackgroundColor, tview.AlignLeft).SetSelectable(false),
	}
}

func (w *workItems) Update() Update {
	size := w.list.Size()
	update := Update{
		Rows:   make([][]*tview.TableCell, 0, size),
		Reload: w.filters.updated(),
	}
	for item := range w.list.All() {
		if row := w.buildRow(item); row != nil {
			update.Rows = append(update.Rows, row)
		}
	}
	// maybe rows were added after get queue.Size()?
	size = max(size, len(update.Rows))
	update.Title = w.title(size, len(update.Rows))

	return update
}

func padString(s string, width int) string {
	if toPad := width - len(s); toPad > 0 {
		s += strings.Repeat(" ", toPad)
	}
	return s
}

func (w *workItems) HandleInput(event *tcell.EventKey) *tcell.EventKey {
	if event.Key() != tcell.KeyRune {
		return event
	}
	if event.Modifiers() != tcell.ModNone {
		return event
	}
	switch event.Rune() {
	case 's':
		w.filters.toggle(pipeline.Skipped)
		return nil
	case 'c':
		w.filters.toggle(pipeline.Converted)
		return nil
	case 'r':
		w.filters.toggle(pipeline.Rejected)
		return nil
	case 'f':
		w.fullName.Store(!w.fullName.Load())
		return nil
	}
	return event
}

func (w *workItems) buildRow(item *pipeline.WorkItem) []*tview.TableCell {
	workStatus := item.WorkStatus()
	if w.filters.on(workStatus.Status) {
		return nil
	}
	source := item.Source.Path
	if !w.fullName.Load() {
		source = filepath.Base(source)
	}
	var errString string
	if err := workStatus.Err; err != nil {
		errString = err.Error()
	}
	return []*tview.TableCell{
		getTableCell(source, tview.Styles.PrimaryTextColor, tview.Styles.PrimitiveBackgroundColor, tview.AlignLeft).SetReference(item),
		getTableCell(item.SourceVideoStats().String(), tview.Styles.PrimaryTextColor, tview.Styles.PrimitiveBackgroundColor, tview.AlignLeft),
		getTableCell(item.TargetVideoStats().String(), tview.Styles.PrimaryTextColor, tview.Styles.PrimitiveBackgroundColor, tview.AlignLeft),
		getTableCell(workStatus.Status.String(), colorStatus(workStatus.Status), tview.Styles.PrimitiveBackgroundColor, tview.AlignLeft),
		getTableCell(item.CompletedFormatted(), tview.Styles.PrimaryTextColor, tview.Styles.PrimitiveBackgroundColor, tview.AlignRight),
		getTableCell(item.RemainingFormatted(), tview.Styles.PrimaryTextColor, tview.Styles.PrimitiveBackgroundColor, tview.AlignRight),
		getTableCell(errString, tview.Styles.PrimaryTextColor, tview.Styles.PrimitiveBackgroundColor, tview.AlignLeft),
	}
}

func (w *workItems) title(itemCount, rowCount int) string {
	title := "files"
	filtered := w.filters.format()
	if filtered != "" {
		title += " (filtered: " + filtered + ")[" + strconv.Itoa(rowCount) + "/" + strconv.Itoa(itemCount) + "]"
	} else {
		title += " [" + strconv.Itoa(rowCount) + "]"
	}
	return " " + title + " "
}
