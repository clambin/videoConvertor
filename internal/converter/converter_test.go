package converter

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/clambin/videoConvertor/internal/configuration"
	"github.com/clambin/videoConvertor/internal/ffmpeg"
	"github.com/clambin/videoConvertor/internal/profile"
	"github.com/clambin/videoConvertor/internal/worklist"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConverter_convert(t *testing.T) {
	type fileCheckerResults struct {
		ok  bool
		err error
	}
	testCases := []struct {
		name               string
		profile            string
		ffmpegErr          error
		fileCheckerResults fileCheckerResults
		want               worklist.WorkStatus
		wantErr            bool
	}{
		{
			name:      "video conversion failed",
			profile:   "hevc-low",
			ffmpegErr: errors.New("failed"),
			want:      worklist.Failed,
			wantErr:   true,
		},
		{
			name:    "video converts successfully",
			profile: "hevc-low",
			want:    worklist.Converted,
			wantErr: false,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var l worklist.WorkList
			l.SetActive(true)

			var cfg configuration.Configuration
			cfg.Profile, _ = profile.GetProfile(tt.profile)

			ff := fakeFFMPEG{err: tt.ffmpegErr}
			c := New(&ff, &l, cfg, slog.Default())
			c.fileChecker = fakeFsChecker{ok: tt.fileCheckerResults.ok, err: tt.fileCheckerResults.err}

			go func() { c.Run(t.Context()) }()

			i := l.Add("foo.mkv")
			i.SetStatus(worklist.Inspected, nil)
			i.AddSourceStats(ffmpeg.VideoStats{VideoCodec: "h264", BitRate: 4_000_000})

			assert.Eventually(t, func() bool {
				status, err := i.Status()
				return status == tt.want && ((tt.wantErr && err != nil) || (!tt.wantErr && err == nil))
			}, time.Second, time.Millisecond)
		})
	}
}

var _ FFMPEG = &fakeFFMPEG{}

type fakeFFMPEG struct {
	err error
}

func (f *fakeFFMPEG) Convert(_ context.Context, _ ffmpeg.Request) error {
	return f.err
}

var _ fileChecker = &fakeFsChecker{}

type fakeFsChecker struct {
	ok  bool
	err error
}

func (f fakeFsChecker) TargetIsNewer(_, _ string) (bool, error) {
	return f.ok, f.err
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
