package main

import (
	"errors"
	"flag"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/KasimKaizer/SweetClover/internal/music"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
)

/* List model */

type tuiMusic struct {
	*music.Music
	width *int
}

func (m *tuiMusic) FilterValue() string {
	return m.Name
}

func (m *tuiMusic) Title() string {
	return truncate(m.Name, fineTuneSize(*m.width, 0.3)) //nolint:gomnd // fineTuning
}

func (m *tuiMusic) Description() string {
	return truncate(m.Artist, fineTuneSize(*m.width, 0.3)) //nolint:gomnd // fineTuning
}

/* End List Model */

/* Main Model */

type Model struct {
	selectedIdx    int
	list           list.Model
	progress       progress.Model
	controller     *music.Controller
	done           chan struct{}
	currentPlaying string
	displayedImg   string
	width          int
	height         int
}

func newModel(path string) (*Model, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	var collection []list.Item

	model := &Model{ //nolint:exhaustruct // its fine to have default values.
		progress: progress.New(
			progress.WithDefaultScaledGradient(),
			progress.WithoutPercentage(),
		),
		controller: music.NewController(),
	}

	if fileInfo.IsDir() {
		collection, err = generateMusicCollection(path, &model.width)
		if err != nil {
			return nil, err
		}
	} else {
		m, mErr := music.NewMusic(path)
		if mErr != nil {
			return nil, mErr
		}
		newMusic := &tuiMusic{
			Music: m,
			width: &model.width,
		}
		collection = append(collection, newMusic)
	}

	if len(collection) == 0 {
		return nil, errors.New("no music files in the provided path")
	}

	model.list = list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	model.list.Title = filepath.Base(path)
	model.list.SetItems(collection)
	return model, nil
}

/* Implement Bubbletea model */

func (m *Model) Init() tea.Cmd {
	return m.progressCmd()
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
			m, musicErr := music.NewMusic(path)
			if err != nil {
				setErr(musicErr)
			}
			newMusic := &tuiMusic{
				Music: m,
				width: textWidth,
			}
			mChan <- newMusic
		}()
		return nil
	}
	_ = filepath.Walk(path, walk) // error ignored as its always nil.
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
	// slices.SortStableFunc(collection, func(a, b list.Item) int {
	// 	return cmp.Compare(a.(*tuiMusic).Name, b.(*tuiMusic).Name)
	// })
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
	if _, err = p.Run(); err != nil {
		log.Fatal(err)
	}
}
