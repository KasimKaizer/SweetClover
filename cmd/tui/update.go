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
		m.style = newStyles(msg.Width, msg.Height)
		m.setHomeScreenOpts()
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k", "down", "j", "right", "left":
			item := m.list.SelectedItem()
			if item != nil {
				m.selected = item.(*tuiMusic)
			}
		case "enter":
			m.displayedScreen = musicScreen
			m.setMusicScreenOpts()
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

func (m *Model) updateMusicScreen(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.style = newStyles(msg.Width, msg.Height)
		m.setMusicScreenOpts()
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyUp:
			m.list.SetShowTitle(false)
		case tea.KeyEsc:
			m.displayedScreen = homeScreen
			m.style = newStyles(m.style.width, m.style.height)
			m.setHomeScreenOpts()
			cmds = append(cmds, lazyLoadImageCmd(
				m.selected,
				fineTuneSize(m.style.height, 0.6),
				fineTuneSize(m.style.width, 0.35),
				m.list.Index(),
			))
		}

	}
	return m, tea.Batch(cmds...)
}

func (m *Model) setMusicScreenOpts() {
	m.list.SetShowTitle(false)
	m.list.SetFilteringEnabled(false)
	m.list.SetShowFilter(false)
	m.list.SetShowStatusBar(false)
	m.list.SetSize(m.style.width, fineTuneSize(m.style.height, 0.5))
	m.progress.Width = fineTuneSize(m.style.width, 0.8)
	m.displayedTextWidth = fineTuneSize(m.style.width, 0.25)
}

func (m *Model) setHomeScreenOpts() {
	m.list.SetShowTitle(true)
	m.list.SetFilteringEnabled(true)
	m.list.SetShowFilter(true)
	m.list.SetShowStatusBar(true)
	m.list.SetSize(m.style.width, m.style.height)
	m.displayedTextWidth = fineTuneSize(m.style.width, 0.3)
}

/* various custom cmd */
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
