package pipeline

import (
	"testing"
	"time"

	"github.com/clambin/videoConvertor/ffmpeg"
	"github.com/clambin/videoConvertor/internal/processor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQueue(t *testing.T) {
	var l Queue

	// default Queue is empty and inactive
	assert.Empty(t, l.List())
	assert.Zero(t, l.Size())
	assert.False(t, l.Active())
	for range l.All() {
		t.Errorf("should not be reached")
	}

	l.Add("foo")
	assert.Equal(t, 1, l.Size())
	var count int
	for range l.All() {
		count++
	}
	assert.Equal(t, l.Size(), count)

}

func TestQueue_NextToConvert(t *testing.T) {
	var l Queue

	// manually waiting items are returned even if the Queue is inactive
	l.Queue(&WorkItem{Source: "foo"})
	i := l.NextToConvert()
	require.NotNil(t, i)
	assert.Equal(t, "foo", i.Source)
	status, err := i.Status()
	assert.Equal(t, Converting, status)
	assert.NoError(t, err)

	// automatically added items are not returned if the Queue is inactive
	l.Add("foo").status = Inspected
	i = l.NextToConvert()
	assert.Nil(t, i)

	// automatically added items are not returned if the Queue is active
	l.SetActive(true)
	i = l.NextToConvert()
	require.NotNil(t, i)
	status, err = i.Status()
	assert.Equal(t, Converting, status)
	assert.NoError(t, err)
}

func TestQueue_Active(t *testing.T) {
	var l Queue
	assert.False(t, l.Active())
	l.SetActive(true)
	assert.True(t, l.Active())
	l.ToggleActive()
	assert.False(t, l.Active())
}

func TestQueue_Stats(t *testing.T) {
	stats := ffmpeg.VideoStats{VideoCodec: "h264", Height: 1080, BitRate: 5_000_000}
	var item WorkItem
	item.AddSourceStats(stats)
	assert.Equal(t, stats, item.SourceVideoStats())
	item.AddTargetStats(stats)
	assert.Equal(t, stats, item.TargetVideoStats())
}

func TestWorkItem_RemainingFormatted(t *testing.T) {
	tests := []struct {
		name     string
		status   WorkStatus
		input    time.Duration
		expected string
	}{
		{"No duration", Converting, 0, ""},
		{"Only seconds", Converting, 10 * time.Second, "10s"},
		{"Only minutes", Converting, 5 * time.Minute, "5m"},
		{"Only hours", Converting, 2 * time.Hour, "2h"},
		{"Days and hours", Converting, 26 * time.Hour, "1d2h"},
		{"Hours and minutes", Converting, 3*time.Hour + 45*time.Minute, "3h45m"},
		{"Hours, minutes, seconds", Converting, 1*time.Hour + 30*time.Minute + 20*time.Second, "1h30m20s"},
		{"Hours and seconds", Converting, 1*time.Hour + 2*time.Second, "1h2s"},
		{"Exactly two days", Converting, 48 * time.Hour, "2d"},
		{"Multiple units", Converting, 72*time.Hour + 10*time.Minute + 15*time.Second, "3d10m15s"},
		{"Not converting", Converted, 72*time.Hour + 10*time.Minute + 15*time.Second, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var item WorkItem
			item.SetStatus(tt.status, nil)
			item.Progress.Duration = 2 * tt.input
			item.Progress.Update(processor.Progress{Converted: tt.input, Speed: 1})
			assert.Equal(t, tt.expected, item.RemainingFormatted())
		})
	}
}

func TestWorkItem_CompletedFormatted(t *testing.T) {
	tests := []struct {
		name   string
		status WorkStatus
		input  time.Duration
		want   string
	}{
		{"not converting", Converted, time.Second, ""},
		{"starting", Converting, 0, ""},
		{"half done", Converting, 30 * time.Minute, "50.0%"},
		{"done", Converting, time.Hour, "100.0%"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var item WorkItem
			item.SetStatus(tt.status, nil)
			item.Progress.Duration = time.Hour
			item.Progress.Update(processor.Progress{Converted: tt.input, Speed: 1})
			assert.Equal(t, tt.want, item.CompletedFormatted())
		})
	}
}

func TestWorkStatus_String(t *testing.T) {
	for val, label := range workStatusToString {
		assert.Equal(t, label, val.String())
	}
	assert.Equal(t, "unknown", WorkStatus(-1).String())
}
