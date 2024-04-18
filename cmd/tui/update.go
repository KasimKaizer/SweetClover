package main

import (
	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) updateHomeScreen(msg tea.Msg) (tea.Model, tea.Cmd) {
	// list's update method
	var cmds []tea.Cmd

	var listCmd tea.Cmd
	m.list, listCmd = m.list.Update(msg)
	cmds = append(cmds, listCmd)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetSize(msg.Width, msg.Height)
		m.style = newStyles(msg.Width, msg.Height)
		m.displayedTextWidth = fineTuneSize(msg.Width, 0.3)
		// _globalTextWidth = fineTuneSize(msg.Width, 0.3)
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k", "down", "j", "right", "left":
			item := m.list.SelectedItem()
			if item != nil {
				m.selected = item.(*tuiMusic)
			}
		}
	case gotImage:
		if msg.idx != m.list.Index() {
			break
		}
		m.displayedImg = msg.image
	}

	// A way to have a loading image
	switch msg.(type) {
	case tea.KeyMsg, tea.WindowSizeMsg:
		m.displayedImg = "LOADING..."
		cmds = append(cmds, lazyLoadImageCmd(
			m.selected,
			fineTuneSize(m.style.height, 0.6),
			fineTuneSize(m.style.width, 0.35),
			m.list.Index(),
		))
	}

	return m, tea.Batch(cmds...)
}

/* End of Bubble Tea required methods */
type gotImage struct {
	image string
	idx   int
}

func lazyLoadImageCmd(music *tuiMusic, height, width, idx int) tea.Cmd {
	return func() tea.Msg {
		img, err := music.GetCoverArtASCII(height, width)
		if err != nil {
			img = "ERROR"
		}
		return gotImage{image: img, idx: idx}
	}
}
