package ui

import (
	"github.com/clambin/videoConvertor/internal/pipeline"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const (
	labelColor    = tcell.ColorSandyBrown
	textColor     = tcell.ColorGrey
	shortcutColor = tcell.ColorBlue
)

var tableColorStatus = map[pipeline.WorkStatus]tcell.Color{
	pipeline.Failed:     tcell.ColorRed,
	pipeline.Rejected:   tcell.ColorRed,
	pipeline.Converted:  tcell.ColorGreen,
	pipeline.Skipped:    tcell.ColorGreen,
	pipeline.Converting: tcell.ColorOrange,
}

func init() {
	tview.Styles.TitleColor = tcell.ColorMediumTurquoise
	tview.Styles.BorderColor = tcell.ColorLightSkyBlue
	tview.Styles.PrimaryTextColor = tcell.ColorLightSkyBlue
	tview.Styles.SecondaryTextColor = tcell.ColorWhite
	tview.Styles.TertiaryTextColor = tcell.ColorGrey
}
