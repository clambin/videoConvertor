package ui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type DataSource interface {
	Update() Update
	HandleInput(event *tcell.EventKey) *tcell.EventKey
}

type Update struct {
	Headers []*tview.TableCell
	Rows    [][]*tview.TableCell
	Title   string
	Reload  bool
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
		SetBorderPadding(0, 0, 1, 1).
		SetInputCapture(t.handleInput)
	return &t
}

func (t *Table) Update() {
	u := t.DataSource.Update()
	for i, h := range u.Headers {
		t.Table.SetCell(0, i, h)
	}
	for i, row := range u.Rows {
		for j, cell := range row {
			t.Table.SetCell(i+1, j, cell)
		}
	}
	t.trimRows(len(u.Rows) + 1)
	if u.Reload {
		t.Table.Select(1, 0)
		t.Table.ScrollToBeginning()
	}
	t.Table.SetTitle(u.Title)
}

func (t *Table) trimRows(rows int) {
	for t.Table.GetRowCount() > rows {
		t.Table.RemoveRow(t.Table.GetRowCount() - 1)
	}
}

func (t *Table) handleInput(event *tcell.EventKey) *tcell.EventKey {
	return t.DataSource.HandleInput(event)
}
