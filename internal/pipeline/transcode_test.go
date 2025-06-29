package pipeline

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/clambin/xcoder/ffmpeg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTranscode(t *testing.T) {
	testCases := []struct {
		name      string
		profile   string
		ffmpegErr error
		want      Status
		wantErr   bool
	}{
		{
			name:      "video conversion failed",
			profile:   "hevc-low",
			ffmpegErr: errors.New("failed"), //nolint:err113
			want:      Failed,
			wantErr:   true,
		},
		{
			name:    "video converts successfully",
			profile: "hevc-low",
			want:    Converted,
			wantErr: false,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			//ff := fakeTranscoder{err: tt.ffmpegErr}
			var q Queue
			q.SetActive(true)
			var cfg Configuration
			cfg.Profile, _ = GetProfile(tt.profile)
			l := slog.New(slog.DiscardHandler)

			go Transcode(t.Context(), &q, cfg, l)

			i := q.Add("foo.mkv")
			i.Source.VideoStats = ffmpeg.VideoStats{VideoCodec: "h264", BitRate: 8_000_000}
			i.Target.VideoStats = ffmpeg.VideoStats{VideoCodec: "hevc", BitRate: 4_000_000}
			i.transcoder = &fakeTranscoder{err: tt.ffmpegErr}
			i.SetWorkStatus(WorkStatus{Status: Inspected})

			assert.Eventually(t, func() bool {
				workStatus := i.WorkStatus()
				return workStatus.Status == tt.want && ((tt.wantErr && workStatus.Err != nil) || (!tt.wantErr && workStatus.Err == nil))
			}, time.Second, convertInterval)
		})
	}
}

func TestFsFileChecker_TargetIsNewer(t *testing.T) {
	tmpDir := t.TempDir()

	filenameA := filepath.Join(tmpDir, "a")
	require.NoError(t, os.WriteFile(filenameA, []byte("A"), 0644))
	require.NoError(t, os.Chtimes(filenameA, time.Now().Add(-time.Hour), time.Now().Add(-time.Hour)))
	filenameB := filepath.Join(tmpDir, "b")
	require.NoError(t, os.WriteFile(filenameB, []byte("B"), 0644))

	tests := []struct {
		name    string
		source  string
		target  string
		wantOK  assert.BoolAssertionFunc
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name:    "newer",
			source:  filenameA,
			target:  filenameB,
			wantOK:  assert.True,
			wantErr: assert.NoError,
		},
		{
			name:    "not newer",
			source:  filenameB,
			target:  filenameA,
			wantOK:  assert.False,
			wantErr: assert.NoError,
		},
		{
			name:    "same is not newer",
			source:  filenameA,
			target:  filenameA,
			wantOK:  assert.False,
			wantErr: assert.NoError,
		},
		{
			name:    "source is missing",
			source:  "invalid",
			target:  filenameB,
			wantOK:  assert.False,
			wantErr: assert.Error,
		},
		{
			name:    "target is missing",
			source:  filenameA,
			target:  "invalid",
			wantOK:  assert.False,
			wantErr: assert.NoError,
		},
	}

	var c fsFileChecker
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ok, err := c.TargetIsNewer(tt.source, tt.target)
			tt.wantOK(t, ok)
			tt.wantErr(t, err)
		})
	}
}

var _ fileChecker = &fakeFsChecker{}

type fakeFsChecker struct {
	ok  bool
	err error
}

func (f fakeFsChecker) TargetIsNewer(_, _ string) (bool, error) {
	return f.ok, f.err
}

var _ transcoder = &fakeTranscoder{}

type fakeTranscoder struct {
	err error
}

func (f *fakeTranscoder) Progress(_ func(ffmpeg.Progress), _ string) *ffmpeg.FFMPEG {
	return nil
}

func (f *fakeTranscoder) Run(_ context.Context, _ *slog.Logger) error {
	return f.err
}
