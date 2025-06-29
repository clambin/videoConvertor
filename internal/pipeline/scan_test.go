package pipeline

import (
	"log/slog"
	"testing"

	"github.com/psanford/memfs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScanFS(t *testing.T) {
	fs := memfs.New()
	require.NoError(t, fs.MkdirAll("foo", 0755))
	require.NoError(t, fs.WriteFile("foo/video.MKV", []byte(""), 0644))
	require.NoError(t, fs.WriteFile("foo/info.txt", []byte(""), 0644))
	require.NoError(t, fs.MkdirAll("bar", 0000))

	var list Queue
	ch := make(chan *WorkItem)
	errCh := make(chan error)
	go func() { errCh <- ScanFS(t.Context(), fs, "/", &list, ch, slog.Default()) }()
	item := <-ch
	assert.Equal(t, "/foo/video.MKV", item.Source.Path)
	require.NoError(t, <-errCh)

	require.Len(t, list.List(), 1)
	assert.Equal(t, "/foo/video.MKV", list.List()[0].Source.Path)
}
