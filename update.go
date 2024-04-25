package main

import (
	"time"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) updateHomeScreen(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	var listCmd tea.Cmd
	m.list, listCmd = m.list.Update(msg)
	cmds = append(cmds, listCmd)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetSize(msg.Width, msg.Height)
		m.progress.Width = fineTuneSize(msg.Width, 0.4)
		m.width, m.height = msg.Width, msg.Height
	case playMusic:
		if !msg.continuePlaying {
			break
		}
		music := m.list.Items()[msg.idx].(*tuiMusic)
		m.currentPlaying = music.Name
		musicCmd := m.playMusicCmd(msg.idx)
		cmds = append(cmds, musicCmd)
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k", "down", "j", "right", "left":
			item := m.list.SelectedItem()
			if item != nil {
				m.selected = item.(*tuiMusic)
			}
		case "p":
			m.controller.PauseResume()
		case "enter":
			if m.done == nil {
				m.done = make(chan struct{})
			} else {
				m.done <- struct{}{}
			}
			m.currentPlaying = m.selected.Name
			cmds = append(cmds, m.playMusicCmd(m.list.Index()))
		}
	case gotImage:
		if msg.idx != m.list.Index() {
			break
		}
		m.displayedImg = msg.image
	case progressStatus:
		cmd := m.progress.SetPercent(float64(msg))
		cmds = append(cmds, cmd, m.progressCmd())
	case progress.FrameMsg:
		progressModel, progCmd := m.progress.Update(msg)
		m.progress = progressModel.(progress.Model)
		cmds = append(cmds, progCmd)
	}

	// A way to have a loading image
	switch msg.(type) {
	case tea.KeyMsg, tea.WindowSizeMsg:
		m.displayedImg = "LOADING..."
		cmds = append(cmds, m.lazyLoadImageCmd(m.list.Index()))
	}

	return m, tea.Batch(cmds...)
}

/* End of Bubble Tea required methods. */
type gotImage struct {
	image string
	idx   int
}

func (m *Model) lazyLoadImageCmd(idx int) tea.Cmd {
	return func() tea.Msg {
		height, width := fineTuneSize(m.height, 0.55), fineTuneSize(m.width, 0.35)
		music := m.list.Items()[idx].(*tuiMusic)
		img, err := music.GetCoverArtASCII(height, width)
		if err != nil {
			img = "ERROR"
		}
		return gotImage{image: img, idx: idx}
	}
}

type playMusic struct {
	idx             int
	continuePlaying bool
	// err error // TODO: implement error handling for this cmd
}

func (m *Model) playMusicCmd(idx int) tea.Cmd {
	return func() tea.Msg {
		list := m.list.Items()
		music := list[idx].(*tuiMusic)
		m.controller.Play(music.Music)
		select {
		case <-m.controller.Done:
			next := (idx + 1) % len(list)
			return playMusic{idx: next, continuePlaying: true}
		case <-m.done:
			return playMusic{idx: 0, continuePlaying: false}
		}
	}
}

type progressStatus float64

func (m *Model) progressCmd() tea.Cmd {
	return tea.Tick(time.Second/2, func(time.Time) tea.Msg {
		prog, err := m.controller.Progress()
		if err != nil {
			prog = 0
		}
		return progressStatus(prog)
	})
}
