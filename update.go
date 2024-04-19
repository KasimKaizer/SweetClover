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
		m.progress.Width = fineTuneSize(msg.Width, 0.4)
		m.width, m.height = msg.Width, msg.Height

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
			fineTuneSize(m.height, 0.55),
			fineTuneSize(m.width, 0.35),
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
