package main

import (
	"cmp"
	"errors"
	"flag"
	"log"
	"os"
	"path/filepath"
	"slices"
	"sync"

	"github.com/KasimKaizer/SweetClover/internal/music"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
)

/* List model */

type tuiMusic struct {
	*music.Music
	displayedTextWidth *int
}

func (m *tuiMusic) FilterValue() string {
	return m.Name
}

func (m *tuiMusic) Title() string {
	return truncate(m.Name, fineTuneSize(*m.displayedTextWidth, 0.3))
}

func (m *tuiMusic) Description() string {
	return truncate(m.Artist, fineTuneSize(*m.displayedTextWidth, 0.3))
}

/* End List Model */

/* Main Model */

type Model struct {
	selected     *tuiMusic
	list         list.Model
	progress     progress.Model
	controller   *music.Controller
	done         chan struct{}
	displayedImg string
	width        int
	height       int
}

func newModel(path string) (*Model, error) {

	fileInfo, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	var collection []list.Item

	model := &Model{
		progress:   progress.New(progress.WithDefaultScaledGradient(), progress.WithoutPercentage()),
		controller: music.NewController(),
	}

	if fileInfo.IsDir() {
		collection, err = generateMusicCollection(path, &model.width)
		if err != nil {
			return nil, err
		}

	} else {
		m, err := music.NewMusic(path)
		if err != nil {
			return nil, err
		}

		newMusic := &tuiMusic{
			Music:              m,
			displayedTextWidth: &model.width,
		}

		collection = append(collection, newMusic)
	}

	if len(collection) == 0 {
		return nil, errors.New("no music files in the provided path")
	}

	model.list = list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	model.list.Title = filepath.Base(path)
	model.list.SetItems(collection)
	model.selected = collection[0].(*tuiMusic)
	return model, nil
}

/* Implement Bubbletea model */

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m.updateHomeScreen(msg)
}

func (m *Model) View() string {
	if m.height == 0 {
		return "Loading..."
	}
	return m.homePageView()
}

func generateMusicCollection(path string, textWidth *int) ([]list.Item, error) {
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
			m, err := music.NewMusic(path)
			if err != nil {
				setErr(err)
			}
			newMusic := &tuiMusic{
				Music:              m,
				displayedTextWidth: textWidth,
			}
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

	// TODO: For some reason sorting doesn't work, check out how list components is
	//       created, and if its possible to sort
	slices.SortStableFunc(collection, func(a, b list.Item) int {
		return cmp.Compare(a.(*tuiMusic).Name, b.(*tuiMusic).Name)
	})
	return collection, nil
}

func main() {
	flag.Parse()
	path := flag.Arg(0)
	// path = ""
	model, err := newModel(path)
	if err != nil {
		log.Fatal(err)
	}
	p := tea.NewProgram(model, tea.WithAltScreen())
	p.SetWindowTitle("Sweet Clover")
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
