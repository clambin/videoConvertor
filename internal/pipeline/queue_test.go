package pipeline

import (
	"testing"
	"time"

	"github.com/clambin/xcoder/ffmpeg"
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
	var queue Queue

	// manually waiting items are returned even if the Queue is inactive
	source := MediaFile{Path: "foo"}
	queue.Queue(&WorkItem{Source: source})
	i := queue.NextToConvert()
	require.NotNil(t, i)
	assert.Equal(t, source, i.Source)
	assert.Equal(t, WorkStatus{Status: Converting}, i.WorkStatus())

	// automatically added items are not returned if the Queue is inactive
	queue.Add("foo").workStatus.Status = Inspected
	i = queue.NextToConvert()
	assert.Nil(t, i)

	// automatically added items are not returned if the Queue is active
	queue.SetActive(true)
	i = queue.NextToConvert()
	require.NotNil(t, i)
	assert.Equal(t, WorkStatus{Status: Converting}, i.WorkStatus())
}

func TestQueue_Active(t *testing.T) {
	var l Queue
	assert.False(t, l.Active())
	l.SetActive(true)
	assert.True(t, l.Active())
	l.ToggleActive()
	assert.False(t, l.Active())
}

func TestWorkItem_RemainingFormatted(t *testing.T) {
	tests := []struct {
		name     string
		status   Status
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
			item.SetWorkStatus(WorkStatus{Status: tt.status})
			item.Progress.Duration = 2 * tt.input
			item.Progress.Update(ffmpeg.Progress{Converted: tt.input, Speed: 1})
			assert.Equal(t, tt.expected, item.RemainingFormatted())
		})
	}
}

func TestWorkItem_CompletedFormatted(t *testing.T) {
	tests := []struct {
		name   string
		status Status
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
			item.SetWorkStatus(WorkStatus{Status: tt.status})
			item.Progress.Duration = time.Hour
			item.Progress.Update(ffmpeg.Progress{Converted: tt.input, Speed: 1})
			assert.Equal(t, tt.want, item.CompletedFormatted())
		})
	}
}

func TestWorkItem_VideoStats(t *testing.T) {
	i := &WorkItem{
		Source: MediaFile{Path: "foo", VideoStats: ffmpeg.VideoStats{VideoCodec: "h264", Height: 1080, BitRate: 8_000_000}},
		Target: MediaFile{Path: "bar", VideoStats: ffmpeg.VideoStats{VideoCodec: "hevc", Height: 1080, BitRate: 4_000_000}},
	}
	assert.Equal(t, "h264/1080/8.00 mbps", i.SourceVideoStats().String())
	assert.Equal(t, "hevc/1080/4.00 mbps", i.TargetVideoStats().String())
}

func TestWorkStatus_String(t *testing.T) {
	tests := []struct {
		want   string
		status Status
	}{
		{"waiting", Waiting},
		{"inspecting", Inspecting},
		{"skipped", Skipped},
		{"inspected", Inspected},
		{"rejected", Rejected},
		{"converting", Converting},
		{"converted", Converted},
		{"failed", Failed},
		{"unknown", -1},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.status.String())
		})
	}
}
