package ui

import (
	"strconv"
	"testing"

	"codeberg.org/clambin/go-common/set"
	"github.com/clambin/videoConvertor/internal/pipeline"
	"github.com/clambin/videoConvertor/internal/profile"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_workListViewer(t *testing.T) {
	var list pipeline.Queue
	list.Add("A").SetStatus(pipeline.Skipped, profile.ErrSourceInTargetCodec)
	list.Add("B").SetStatus(pipeline.Rejected, profile.ErrSourceRejected{Reason: "bitrate too low"})
	list.Add("C").SetStatus(pipeline.Inspected, nil)
	list.Add("D").SetStatus(pipeline.Converted, nil)

	tests := []struct {
		name      string
		filters   []pipeline.WorkStatus
		wantCount int
		wantFirst string
	}{
		{
			name:      "no filters",
			wantCount: 1 + len(list.List()),
			wantFirst: "A",
		},
		{
			name:      "filter skipped",
			filters:   []pipeline.WorkStatus{pipeline.Skipped},
			wantCount: 1 + len(list.List()) - 1,
			wantFirst: "B",
		},
		{
			name:      "filter rejected",
			filters:   []pipeline.WorkStatus{pipeline.Rejected},
			wantCount: 1 + len(list.List()) - 1,
			wantFirst: "A",
		},
		{
			name:      "filter skipped & rejected",
			filters:   []pipeline.WorkStatus{pipeline.Skipped, pipeline.Rejected},
			wantCount: 1 + len(list.List()) - 2,
			wantFirst: "C",
		},
		{
			name:      "all filter skipped & rejected",
			filters:   []pipeline.WorkStatus{pipeline.Skipped, pipeline.Rejected, pipeline.Inspected, pipeline.Converted},
			wantCount: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := newQueueViewer(&list)
			v.DataSource.(*workItems).filters.toggle(tt.filters...)
			v.refresh()

			assert.Equal(t, tt.wantCount, v.GetRowCount())
			if tt.wantFirst != "" {
				require.Greater(t, v.GetRowCount(), 1)
				assert.Equal(t, tt.wantFirst, v.GetCell(1, 0).Text)
			}
		})
	}

}

func Test_workItems_title(t *testing.T) {
	tests := []struct {
		name          string
		filters       []pipeline.WorkStatus
		itemCount     int
		rowCount      int
		expectedTitle string
	}{
		{
			name:          "No filter, multiple items",
			itemCount:     10,
			rowCount:      5,
			expectedTitle: " files [5] ",
		},
		{
			name:          "With filter, multiple items",
			filters:       []pipeline.WorkStatus{pipeline.Skipped},
			itemCount:     10,
			rowCount:      5,
			expectedTitle: " files (filtered: skipped)[5/10] ",
		},
		{
			name:          "With filter, single item",
			filters:       []pipeline.WorkStatus{pipeline.Skipped, pipeline.Rejected},
			itemCount:     1,
			rowCount:      1,
			expectedTitle: " files (filtered: rejected, skipped)[1/1] ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ds := workItems{filters: filters{statuses: set.Set[pipeline.WorkStatus]{}}}
			ds.filters.toggle(tt.filters...)
			assert.Equal(t, tt.expectedTitle, ds.title(tt.itemCount, tt.rowCount))
		})
	}
}

// Current:
// Benchmark_workItems_Update-16               2130            523429 ns/op          105566 B/op       1016 allocs/op
func Benchmark_workItems_Update(b *testing.B) {
	var list pipeline.Queue
	for i := range 1000 {
		list.Add(strconv.Itoa(i))
	}
	updater := workItems{list: &list}
	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		u := updater.Update()
		for _, r := range u.Rows {
			for _, c := range r {
				putTableCell(c)
			}
		}
	}
}
