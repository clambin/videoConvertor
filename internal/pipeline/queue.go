package pipeline

import (
	"iter"
	"slices"
	"strconv"
	"sync"
	"time"

	"github.com/clambin/videoConvertor/ffmpeg"
)

type Queue struct {
	queue   []*WorkItem
	waiting []*WorkItem
	lock    sync.RWMutex
	active  bool
}

func (q *Queue) Add(filename string) *WorkItem {
	q.lock.Lock()
	defer q.lock.Unlock()
	item := &WorkItem{Source: filename}
	q.queue = append(q.queue, item)
	return item
}

func (q *Queue) NextToConvert() *WorkItem {
	// convert any items the user manually asked to convert
	if item := q.dequeue(); item != nil {
		return item
	}
	// is the queue active?
	if !q.Active() {
		return nil
	}
	// return the next item ready for conversion
	return q.checkout(Inspected, Converting)
}

func (q *Queue) dequeue() *WorkItem {
	q.lock.Lock()
	defer q.lock.Unlock()
	var item *WorkItem
	if len(q.waiting) > 0 {
		item = q.waiting[0]
		item.SetStatus(Converting, nil)
		q.waiting = q.waiting[1:]
	}
	return item
}

func (q *Queue) checkout(current, next WorkStatus) *WorkItem {
	q.lock.RLock()
	defer q.lock.RUnlock()
	for _, item := range q.queue {
		if status, _ := item.Status(); status == current {
			item.SetStatus(next, nil)
			return item
		}
	}
	return nil
}

// Queue adds an item ready to be converted. This item will be processed, regardless of whether the queue is active or not.
func (q *Queue) Queue(item *WorkItem) {
	q.lock.Lock()
	defer q.lock.Unlock()
	q.waiting = append(q.waiting, item)
}

// List returns all items in the queue. This clones the contained slice. For performance reasons,
// this should only be used for testing. Use All(), which returns an iterator, instead.
func (q *Queue) List() []*WorkItem {
	q.lock.RLock()
	defer q.lock.RUnlock()
	return slices.Clone(q.queue)
}

func (q *Queue) Size() int {
	q.lock.RLock()
	defer q.lock.RUnlock()
	return len(q.queue)
}

func (q *Queue) All() iter.Seq[*WorkItem] {
	return func(yield func(*WorkItem) bool) {
		q.lock.RLock()
		defer q.lock.RUnlock()
		for _, item := range q.queue {
			if !yield(item) {
				return
			}
		}
	}
}

func (q *Queue) Active() bool {
	q.lock.RLock()
	defer q.lock.RUnlock()
	return q.active
}

func (q *Queue) SetActive(active bool) {
	q.lock.Lock()
	defer q.lock.Unlock()
	q.active = active
}

func (q *Queue) ToggleActive() {
	q.lock.Lock()
	defer q.lock.Unlock()
	q.active = !q.active
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type WorkStatus int

const (
	Waiting WorkStatus = iota
	Inspecting
	Skipped
	Rejected
	Inspected
	Converting
	Converted
	Failed
)

var workStatusToString = map[WorkStatus]string{
	Waiting:    "waiting",
	Inspecting: "inspecting",
	Skipped:    "skipped",
	Rejected:   "rejected",
	Inspected:  "inspected",
	Converting: "converting",
	Converted:  "converted",
	Failed:     "failed",
}

func (ws WorkStatus) String() string {
	if label, ok := workStatusToString[ws]; ok {
		return label
	}
	return "unknown"
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type WorkItem struct {
	err         error
	Source      string
	sourceStats ffmpeg.VideoStats
	targetStats ffmpeg.VideoStats
	Progress    Progress
	status      WorkStatus
	lock        sync.RWMutex
}

func (w *WorkItem) Status() (WorkStatus, error) {
	w.lock.RLock()
	defer w.lock.RUnlock()
	return w.status, w.err
}

func (w *WorkItem) SetStatus(status WorkStatus, err error) {
	w.lock.Lock()
	defer w.lock.Unlock()
	w.status = status
	w.err = err
}

func (w *WorkItem) SourceVideoStats() ffmpeg.VideoStats {
	w.lock.RLock()
	defer w.lock.RUnlock()
	return w.sourceStats
}

func (w *WorkItem) AddSourceStats(stats ffmpeg.VideoStats) {
	w.lock.Lock()
	defer w.lock.Unlock()
	w.sourceStats = stats
	w.Progress.Duration = stats.Duration
}

func (w *WorkItem) TargetVideoStats() ffmpeg.VideoStats {
	w.lock.RLock()
	defer w.lock.RUnlock()
	return w.targetStats
}

func (w *WorkItem) AddTargetStats(stats ffmpeg.VideoStats) {
	w.lock.Lock()
	defer w.lock.Unlock()
	w.targetStats = stats
}

func (w *WorkItem) RemainingFormatted() string {
	w.lock.RLock()
	defer w.lock.RUnlock()
	if w.status != Converting {
		return ""
	}
	var output string
	if d := w.Progress.Remaining(); d >= 0 { // not sure why I added this check?
		output = formatDuration(d)
	}
	return output
}

func formatDuration(d time.Duration) string {
	var output string
	var days int
	if days = int(d.Hours()) / 24; days > 0 {
		output = strconv.Itoa(days) + "d"
		d -= time.Duration(days) * 24 * time.Hour
	}
	if hours := int(d.Hours()); hours > 0 {
		output += strconv.Itoa(hours) + "h"
		d -= time.Duration(hours) * time.Hour
	}
	if minutes := int(d.Minutes()); minutes > 0 {
		output += strconv.Itoa(minutes) + "m"
		d -= time.Duration(minutes) * time.Minute
	}
	if seconds := int(d.Seconds()); seconds > 0 {
		output += strconv.Itoa(seconds) + "s"
	}
	return output
}

func (w *WorkItem) CompletedFormatted() string {
	w.lock.RLock()
	defer w.lock.RUnlock()
	if w.status != Converting {
		return ""
	}
	if p := w.Progress.Completed(); p > 0 {
		return strconv.FormatFloat(100*p, 'f', 1, 64) + "%"
	}
	return ""
}
