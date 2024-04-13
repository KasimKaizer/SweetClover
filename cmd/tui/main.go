package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/dhowden/tag"
)

/* music Model */
type music struct {
	FilePath string
	Name     string
	Artist   string
	// err      error // will be useful later on
}

// we need to implement list.item interface for bubbletea lists
func (m *music) FilterValue() string {
	return m.Name
}

func (m *music) Title() string {
	return m.Name
}

func (m *music) Description() string {
	return m.Artist
}

func (m *music) setData() error {
	f, err := os.Open(m.FilePath)
	if err != nil {
		return err
	}
	defer f.Close()
	fMeta, err := tag.ReadFrom(f)
	if err != nil {
		if errors.Is(err, tag.ErrNoTagsFound) {
			name := filepath.Base(f.Name())
			m.Name = strings.TrimSuffix(name, filepath.Ext(name))
			m.Artist = "UNKNOWN"
			return nil
		}
		return err
	}
	m.Name = fMeta.Title()
	m.Artist = fMeta.Artist()
	return nil
}

/* Main model */
type Model struct {
	path string
	list list.Model
	// err  error // will be useful later on
}

func newModel(path string) *Model {
	return &Model{path: path}
}

func (m *Model) initList(path string, width, height int) error {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return err
	}
	music := []list.Item{&music{FilePath: path}}
	if fileInfo.IsDir() {
		music, err = processDir(path)
		if err != nil {
			return err
		}
	}
	m.list = list.New([]list.Item{}, list.NewDefaultDelegate(), width, height)
	m.list.Title = filepath.Base(path)
	m.list.SetItems(music)
	return nil
}

func processDir(path string) ([]list.Item, error) {
	var collection []list.Item
	walk := func(path string, f os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if f.IsDir() || !isMusic(path) {
			return nil
		}
		newMusic := &music{
			FilePath: path,
		}
		err = newMusic.setData()
		if err != nil {
			return err
		}
		collection = append(collection, newMusic)
		return nil
	}
	err := filepath.Walk(path, walk)
	if err != nil {
		return nil, err
	}
	return collection, nil
}

func isMusic(file string) bool {
	switch filepath.Ext(file) {
	case ".wav", ".mp3", ".ogg", ".flac":
		return true
	}
	return false
}

/* Implement Bubbletea model */

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.initList(m.path, msg.Width, msg.Height)
	}
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m *Model) View() string {
	return m.list.View()
}

func main() {
	p := tea.NewProgram(newModel("test_files"))
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
