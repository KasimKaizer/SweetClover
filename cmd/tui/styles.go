package main

import (
	"fmt"
	"strconv"

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
)

func newStyles(width, height int) *styles {

	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#9B9B9B")).
		PaddingLeft(fineTuneSize(width, 0.075)).
		Width(fineTuneSize(width, 0.5))

	textStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Width(fineTuneSize(width, 0.5))

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

func (m *Model) formatMetaData(img string) string {

	// maxSentenceWidth := fineTuneSize(m.style.width, 0.3)

	return lipgloss.NewStyle().
		Align(lipgloss.Left, lipgloss.Center).
		PaddingTop(fineTuneSize(m.style.height, 0.05)).
		Render(
			// image rendering options
			lipgloss.Place(
				fineTuneSize(m.style.width, 0.5),
				fineTuneSize(m.style.height, 0.7),
				lipgloss.Center, lipgloss.Center,
				m.style.imageStyle.Render(img),
			),
			// music metadata rendering options
			m.style.titleStyle.Render(
				fmt.Sprintf(
					_infoBoxTmpl,
					m.style.textStyle.Render(truncate(m.selected.Name, m.displayedTextWidth)),
					m.style.textStyle.Render(truncate(m.selected.Album, m.displayedTextWidth)),
					m.style.textStyle.Render(truncate(m.selected.Artist, m.displayedTextWidth)),
					m.style.textStyle.Render(strconv.Itoa(m.selected.ReleaseYear))),
			),
		)
}

func (m *Model) homePageView() string {

	// img, err := m.selected.GetCoverArtASCII(
	// 	fineTuneSize(m.style.height, 0.6),
	// 	fineTuneSize(m.style.width, 0.35))

	// if err != nil {
	// 	img = "Error"
	// }

	return lipgloss.JoinHorizontal(lipgloss.Left,
		lipgloss.Place(m.style.width/2,
			m.style.height,
			lipgloss.Top,
			lipgloss.Left,
			m.list.View(),
		),
		m.formatMetaData(m.displayedImg),
	)
}

func truncate(str string, width int) string {
	if width > 3 && len(str) > width {
		str = fmt.Sprintf("%s...", str[:width-3])
	}
	return str
}

func fineTuneSize(num int, deci float64) int {
	return int(float64(num) * deci)
}
