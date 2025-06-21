package pipeline

import (
	"sync"
	"time"

	"github.com/clambin/videoConvertor/internal/processor"
)

type Progress struct {
	lock     sync.RWMutex
	Duration time.Duration
	progress processor.Progress
}

func (p *Progress) Update(progress processor.Progress) {
	p.lock.Lock()
	defer p.lock.Unlock()
	p.progress.Converted = progress.Converted
	// if speed is zero, we won't be able to calculate the remaining time. in this case, we don't update and the
	// remaining time will be calculated using the last  reported speed.
	if progress.Speed != 0 {
		p.progress.Speed = progress.Speed
	}
}

func (p *Progress) Completed() float64 {
	p.lock.RLock()
	defer p.lock.RUnlock()
	return float64(p.progress.Converted) / float64(p.Duration)
}

func (p *Progress) Remaining() time.Duration {
	p.lock.RLock()
	defer p.lock.RUnlock()
	if p.progress.Speed == 0 {
		return -1
	}
	return time.Duration(float64(p.Duration-p.progress.Converted) / p.progress.Speed)
}
