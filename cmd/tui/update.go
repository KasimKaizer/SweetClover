package main

import tea "github.com/charmbracelet/bubbletea"

func (m *Model) updateHomeScreen(msg tea.Msg) (tea.Model, tea.Cmd) {
	// list's update method
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetSize(msg.Width, msg.Height)
		m.style = newStyles(msg.Width, msg.Height)
		m.displayedTextWidth = fineTuneSize(msg.Width, 0.3)
		// _globalTextWidth = fineTuneSize(msg.Width, 0.3)
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k", "down", "j", "right", "left":
			m.selected = m.list.SelectedItem().(*tuiMusic)
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
		cmd = tea.Batch(cmd, tea.Batch(cmd, lazyLoadImage(
			m.selected,
			fineTuneSize(m.style.height, 0.6),
			fineTuneSize(m.style.width, 0.35),
			m.list.Index(),
		)))
	}

	return m, cmd
}
