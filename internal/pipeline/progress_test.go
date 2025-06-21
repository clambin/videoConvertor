package pipeline

import (
	"testing"
	"time"

	"github.com/clambin/videoConvertor/internal/processor"
	"github.com/stretchr/testify/assert"
)

func TestProgress(t *testing.T) {
	tests := []struct {
		name          string
		duration      time.Duration
		progress      processor.Progress
		prevSpeed     float64
		wantCompleted float64
		wantRemaining time.Duration
	}{
		{
			name:          "not initialized",
			duration:      time.Hour,
			wantCompleted: 0,
			wantRemaining: -1,
		},
		{
			name:          "half done",
			duration:      time.Hour,
			progress:      processor.Progress{Speed: 1, Converted: 30 * time.Minute},
			wantCompleted: .5,
			wantRemaining: 30 * time.Minute,
		},
		{
			name:          "speed matters",
			duration:      time.Hour,
			progress:      processor.Progress{Speed: 2, Converted: 30 * time.Minute},
			wantCompleted: .5,
			wantRemaining: 15 * time.Minute,
		},
		{
			name:          "completed",
			duration:      time.Hour,
			progress:      processor.Progress{Speed: 2, Converted: time.Hour},
			wantCompleted: 1,
			wantRemaining: 0,
		},
		{
			name:          "zero speed",
			duration:      time.Hour,
			progress:      processor.Progress{Converted: 30 * time.Minute},
			prevSpeed:     2,
			wantCompleted: .5,
			wantRemaining: 15 * time.Minute,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			p := Progress{Duration: tt.duration}
			p.progress.Speed = tt.prevSpeed
			p.Update(tt.progress)
			assert.Equal(t, tt.wantCompleted, p.Completed())
			assert.Equal(t, tt.wantRemaining, p.Remaining())
		})
	}
}
