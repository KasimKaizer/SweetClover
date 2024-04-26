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
		m.progress.Width = fineTuneSize(msg.Width, 0.4) //nolint:gomnd // fineTuning
		m.width, m.height = msg.Width, msg.Height
		m.displayedImg = "LOADING..."
		cmds = append(cmds, m.lazyLoadImageCmd(m.selectedIdx))
	case playMusic:
		if !msg.continuePlaying {
			break
		}
		music, _ := m.list.VisibleItems()[msg.idx].(*tuiMusic) // will always be *tuiMusic
		m.currentPlaying = music.Name
		musicCmd := m.playMusicCmd(msg.idx)
		cmds = append(cmds, musicCmd)
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k", "down", "j", "right", "left":
			m.selectedIdx = m.list.Index()
			m.displayedImg = "LOADING..."
			cmds = append(cmds, m.lazyLoadImageCmd(m.selectedIdx))
		case "o":
			m.controller.PauseResume()
		case "p":
			_ = m.controller.SeekSeconds(10) // TODO: DISPLAY ERROR TO USER
		case "i":
			_ = m.controller.SeekSeconds(-10) // TODO: DISPLAY ERROR TO USER
		case "enter":
			if m.done == nil {
				m.done = make(chan struct{})
			} else {
				m.done <- struct{}{}
			}
			m.currentPlaying = m.list.VisibleItems()[m.selectedIdx].(*tuiMusic).Name
			cmds = append(cmds, m.playMusicCmd(m.selectedIdx))
		}
	case gotImage:
		if msg.idx != m.selectedIdx {
			break
		}
		m.displayedImg = msg.image
	case progressStatus:
		cmd := m.progress.SetPercent(float64(msg))
		cmds = append(cmds, cmd, m.progressCmd())
	case progress.FrameMsg:
		progressModel, progCmd := m.progress.Update(msg)
		m.progress, _ = progressModel.(progress.Model)
		cmds = append(cmds, progCmd)
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
		height := fineTuneSize(m.height, 0.55)             //nolint:gomnd // fineTuning
		width := fineTuneSize(m.width, 0.35)               //nolint:gomnd // fineTuning
		music, _ := m.list.VisibleItems()[idx].(*tuiMusic) // its will always be *TUIMusic
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
		list := m.list.VisibleItems()
		music, _ := list[idx].(*tuiMusic)
		_ = m.controller.Play(music.Music) // TODO: implement error checking here.
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
	return tea.Tick(time.Second/2, func(time.Time) tea.Msg { //nolint:gomnd // fineTuning
		prog, err := m.controller.Progress()
		if err != nil {
			prog = 0
		}
		return progressStatus(prog)
	})
}
