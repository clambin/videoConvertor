package ui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"sync"
)

type DataSource interface {
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
	DataSource
}

func NewTable(source DataSource) *Table {
	t := Table{
		Table:      tview.NewTable(),
		DataSource: source,
	}
	t.Table.SetEvaluateAllRows(true).
		SetFixed(1, 0).
		SetSelectable(true, false).
		Select(1, 0).
		SetBorder(true).
		SetBorderPadding(0, 0, 1, 1)
	return &t
}

func (t *Table) Update() {
	if t.Table.GetRowCount() == 0 {
		for i, h := range t.DataSource.Header() {
			t.Table.SetCell(0, i, h)
		}
	}
	update := t.DataSource.Update()
	t.Table.SetTitle(update.Title)
	for i, row := range update.Rows {
		for j, cell := range row {
			putTableCell(t.Table.GetCell(i+1, j))
			t.Table.SetCell(i+1, j, cell)
		}
	}
	t.trimRows(len(update.Rows) + 1)
	if update.Reload {
		t.Table.Select(1, 0)
		t.Table.ScrollToBeginning()
	}
}

func (t *Table) trimRows(rows int) {
	for t.Table.GetRowCount() > rows {
		r := t.Table.GetRowCount() - 1
		for c := range t.Table.GetColumnCount() {
			putTableCell(t.Table.GetCell(r, c))
		}
		t.Table.RemoveRow(r)
	}
}

//////////////////////////////////////////////////////////////////////////////////////////////////////

// cellPool reduces memory allocations for tview.TableCell objects
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
