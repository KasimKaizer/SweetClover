//nolint:gomnd // Styles page contains many elements which are relatively sized thus magic numbers are inevitable.
package main

import (
	"fmt"
	"strconv"

	"github.com/charmbracelet/lipgloss"
)

const (
	_infoBoxTmpl = `
Name: %s

Album: %s

Artist: %s

ReleaseYear: %s
`
)

func (m *Model) formatMetaData(img string) string {
	music, _ := m.list.VisibleItems()[m.selectedIdx].(*tuiMusic)

	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#9B9B9B")).
		PaddingTop(fineTuneSize(m.height, 0.03)).
		PaddingLeft(fineTuneSize(m.width, 0.075)).
		Width(fineTuneSize(m.width, 0.5))

	textStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Width(fineTuneSize(m.width, 0.5))

	imgStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(
			lipgloss.Color("#3C3C3C"),
		)
	return lipgloss.NewStyle().
		Align(lipgloss.Left, lipgloss.Center).
		PaddingTop(fineTuneSize(m.height, 0.03)).
		Render(
			// image rendering options
			lipgloss.Place(
				fineTuneSize(m.width, 0.5),
				fineTuneSize(m.height, 0.6),
				lipgloss.Center, lipgloss.Center,
				imgStyle.Render(img),
			),
			// music metadata rendering options
			titleStyle.Render(
				fmt.Sprintf(
					_infoBoxTmpl,
					textStyle.Render(truncate(music.Name, fineTuneSize(m.width, 0.3))),
					textStyle.Render(truncate(music.Album, fineTuneSize(m.width, 0.3))),
					textStyle.Render(truncate(music.Artist, fineTuneSize(m.width, 0.3))),
					textStyle.Render(strconv.Itoa(music.ReleaseYear))),
			),
		)
}

func (m *Model) homePageView() string {
	playingText := truncate(m.currentPlaying, fineTuneSize(m.width, 0.3))
	playingTextPadding := fineTuneSize(fineTuneSize(m.width, 0.5)-lipgloss.Width(playingText), 0.5)

	return lipgloss.JoinHorizontal(lipgloss.Left,
		lipgloss.Place(
			fineTuneSize(m.width, 0.5),
			m.height,
			lipgloss.Top,
			lipgloss.Left,
			m.list.View(),
		),
		lipgloss.JoinVertical(lipgloss.Top,
			lipgloss.NewStyle().
				PaddingTop(fineTuneSize(m.height, 0.04)).
				Foreground(lipgloss.Color("#9B9B9B")).
				PaddingLeft(playingTextPadding).
				Width(fineTuneSize(m.width, 0.5)).
				Render(playingText),
			lipgloss.NewStyle().
				PaddingTop(fineTuneSize(m.height, 0.03)).
				PaddingLeft(fineTuneSize(m.width, 0.05)).
				Render(m.progress.View()),
			m.formatMetaData(m.displayedImg),
		),
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
