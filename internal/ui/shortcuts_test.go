package ui

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_shortcuts(t *testing.T) {
	keys := shortcutsPage{
		{{"a", "key a"}},
		{{"b", "key b"}},
	}

	v := newShortcutsView()
	v.addPage("foo", keys, true)

	require.Equal(t, 1, v.GetPageCount())
	name, _ := v.GetFrontPage()
	assert.Equal(t, "foo", name)
	// TODO: how to get the contents of the Page?
}
