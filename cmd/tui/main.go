package main

import (
	"cmp"
	"flag"
	"log"
	"os"
	"path/filepath"
	"slices"
	"sync"

	"github.com/KasimKaizer/SweetClover/internal/music"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

var (
	_globalTextWidth = 0 // TODO: figure a way out where we won't need this
)

/* List model */

type tuiMusic struct {
	*music.Music
}

func (m *tuiMusic) FilterValue() string {
	return m.Name
}

func (m *tuiMusic) Title() string {
	return truncate(m.Name, _globalTextWidth)
}

func (m *tuiMusic) Description() string {
	return truncate(m.Artist, _globalTextWidth)
}

/* End List Model */

/* Main Model */

type Model struct {
	selected *tuiMusic
	list     list.Model
	style    *styles
}

func newModel(path string) (*Model, error) {
	model := &Model{style: newStyles(0, 0)}
	err := model.initList(path)
	return model, err
}

func (m *Model) initList(path string) error {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return err
	}
	var collection []list.Item

	if fileInfo.IsDir() {
		collection, err = processDir(path)
		if err != nil {
			return err
		}
	} else {
		newMusic := &tuiMusic{
			Music: &music.Music{FilePath: path},
		}
		err := newMusic.PopulateMusicMeta()
		if err != nil {
			return err
		}
		collection = append(collection, newMusic)
	}

	m.list = list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	m.list.Title = filepath.Base(path)
	m.list.SetItems(collection)
	m.selected = collection[0].(*tuiMusic)
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
		m.style = newStyles(msg.Width, msg.Height)
		_globalTextWidth = fineTuneSize(msg.Width, 0.3)
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k", "down", "j", "right", "left":
			m.selected = m.list.SelectedItem().(*tuiMusic)
		}
	}
	return m, cmd
}

func (m *Model) View() string {
	if m.style.height == 0 {
		return "Loading..."
	}
	return m.homePageView()
}

/* End of Bubble Tea required methods */

func processDir(path string) ([]list.Item, error) {
	var collection []list.Item
	var wg sync.WaitGroup
	mChan := make(chan *tuiMusic)

	var onceErr error
	var once sync.Once
	setErr := func(err error) {
		if err != nil {
			once.Do(func() { onceErr = err })
		}
	}
	walk := func(path string, f os.FileInfo, err error) error {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if f.IsDir() || !music.IsMusic(path) {
				return
			}
			newMusic := &tuiMusic{
				Music: &music.Music{FilePath: path},
			}
			setErr(newMusic.PopulateMusicMeta())
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
		if onceErr != nil {
			return nil, onceErr
		}
		collection = append(collection, item)
	}
	slices.SortFunc(collection, func(a, b list.Item) int {
		return cmp.Compare(a.(*tuiMusic).Name, b.(*tuiMusic).Name)
	})
	return collection, nil
}

func main() {
	flag.Parse()
	path := flag.Arg(0)
	path = "/Users/kaizersuterwala/projects_go/sweet_clover/test_files"
	model, err := newModel(path)
	if err != nil {
		log.Fatal(err)
	}
	p := tea.NewProgram(model)
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
