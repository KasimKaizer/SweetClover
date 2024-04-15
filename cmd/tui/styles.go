package main

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

type styles struct {
	titleStyle lipgloss.Style
	textStyle  lipgloss.Style
	imageStyle lipgloss.Style
	width      int
	height     int
}

const (
	_infoBoxTmpl = `
Name: %s

Album: %s

Artist: %s

ReleaseYear: %s
`
	_defaultName = "UNKNOWN"
	_defaultImg  = "NO_IMAGE"
	_defaultYear = 2100
)

func newStyles(width, height int) *styles {
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#9B9B9B")).
		PaddingLeft(width / 13).Width(width / 2)
	textStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).Width(width / 2)

	imgStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(
			lipgloss.Color(lipgloss.Color("#3C3C3C")),
		)
	return &styles{
		titleStyle: titleStyle,
		textStyle:  textStyle,
		imageStyle: imgStyle,
		width:      width,
		height:     height,
	}
}

func truncate(str string, width int) string {
	if width > 3 && len(str) > width {
		str = fmt.Sprintf("%s...", str[:width-3])
	}
	return str
}
