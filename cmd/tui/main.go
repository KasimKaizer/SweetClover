package main

import (
	"cmp"
	"flag"
	"log"
	"os"
	"path/filepath"
	"slices"
	"sync"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

/* Main model */

type Model struct {
	selected *MusicInfoModel
	list     list.Model
	err      error // will be useful later on
}

func newModel(path string) *Model {
	model := &Model{selected: createMusicInfo(0, 0)}
	model.initList(path)
	return model
}

func (m *Model) initList(path string) error {
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
	m.list = list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	m.list.Title = filepath.Base(path)
	m.list.SetItems(music)
	return nil
}

/* Implement Bubbletea model */

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetSize(msg.Width, msg.Height)
		m.selected = createMusicInfo(msg.Width, msg.Height)
		music := m.list.SelectedItem().(*music)
		m.selected.getMusicMeta(music)
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "w", "down", "s":
			music := m.list.SelectedItem().(*music)
			m.selected.getMusicMeta(music)
		}
	}
	return m, cmd
}

func (m *Model) View() string {

	return lipgloss.JoinHorizontal(lipgloss.Left,
		lipgloss.Place(m.selected.style.width/2,
			m.selected.style.height,
			lipgloss.Top,
			lipgloss.Left,
			m.list.View(),
		),
		m.selected.format(),
	)
}

func processDir(path string) ([]list.Item, error) {
	var collection []list.Item
	var wg sync.WaitGroup
	mChan := make(chan *music)
	walk := func(path string, f os.FileInfo, err error) error {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if f.IsDir() || !isMusic(path) {
				return
			}
			newMusic := &music{
				FilePath: path,
			}
			newMusic.err = newMusic.setData()
			mChan <- newMusic
		}()
		return nil // error will always be nil
	}
	filepath.Walk(path, walk)
	go func() {
		wg.Wait()
		close(mChan)
	}()
	for item := range mChan {
		if item.err != nil {
			return nil, item.err
		}
		collection = append(collection, item)
	}
	slices.SortFunc(collection, func(a, b list.Item) int {
		return cmp.Compare(a.FilterValue(), b.FilterValue())
	})
	return collection, nil
}

func main() {
	flag.Parse()
	path := flag.Arg(0)
	path = "/Users/kaizersuterwala/projects_go/sweet_clover/test_files"
	p := tea.NewProgram(newModel(path))
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
