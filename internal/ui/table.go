package ui

import (
	"sync"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type dataSource interface {
	Header() []*tview.TableCell
	Update() Update
	HandleInput(event *tcell.EventKey) *tcell.EventKey
}

type Update struct {
	Title  string
	Rows   [][]*tview.TableCell
	Reload bool
}

type Table struct {
	*tview.Table
	dataSource
}

func NewTable(source dataSource) *Table {
	t := Table{
		Table:      tview.NewTable(),
		dataSource: source,
	}
	t.SetEvaluateAllRows(true).
		SetFixed(1, 0).
		SetSelectable(true, false).
		Select(1, 0).
		SetBorder(true).
		SetBorderPadding(0, 0, 1, 1)
	return &t
}

func (t *Table) Update() {
	if t.GetRowCount() == 0 {
		for i, h := range t.Header() {
			t.SetCell(0, i, h)
		}
	}
	update := t.dataSource.Update()
	t.SetTitle(update.Title)
	for i, row := range update.Rows {
		for j, cell := range row {
			putTableCell(t.GetCell(i+1, j))
			t.SetCell(i+1, j, cell)
		}
	}
	t.trimRows(len(update.Rows) + 1)
	if update.Reload {
		t.Select(1, 0)
		t.ScrollToBeginning()
	}
}

func (t *Table) trimRows(rows int) {
	for t.GetRowCount() > rows {
		r := t.GetRowCount() - 1
		for c := range t.GetColumnCount() {
			putTableCell(t.GetCell(r, c))
		}
		t.RemoveRow(r)
	}
}

//////////////////////////////////////////////////////////////////////////////////////////////////////

// cellPool reduces memory allocations for tview.TableCell objects.
var cellPool = sync.Pool{
	New: func() any {
		return tview.NewTableCell("")
	},
}

func getTableCell(label string, fgColor, bgColor tcell.Color, align int) *tview.TableCell {
	cell := cellPool.Get().(*tview.TableCell)
	cell.Style = tcell.StyleDefault.Foreground(fgColor).Background(bgColor)
	cell.SetText(label).SetAlign(align).SetReference(nil)
	return cell
}

func putTableCell(cell *tview.TableCell) {
	cellPool.Put(cell)
}
