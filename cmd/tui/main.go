package main

import (
	"cmp"
	"flag"
	"log"
	"os"
	"path/filepath"
	"slices"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

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
	slices.SortFunc(collection, func(a, b list.Item) int {
		return cmp.Compare(a.FilterValue(), b.FilterValue())
	})
	return collection, nil
}

func main() {
	flag.Parse()
	path := flag.Arg(0)
	p := tea.NewProgram(newModel(path))
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
