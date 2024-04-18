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

type screen int

const (
	homeScreen screen = iota
	musicScreen
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
	return truncate(m.Name, *m.displayedTextWidth)
}

func (m *tuiMusic) Description() string {
	return truncate(m.Artist, *m.displayedTextWidth)
}

/* End List Model */

/* Main Model */

type Model struct {
	selected           *tuiMusic
	list               list.Model
	style              *styles
	progress           progress.Model
	displayedScreen    screen
	displayedImg       string
	displayedTextWidth int
}

func newModel(path string) (*Model, error) {

	fileInfo, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	var collection []list.Item
	model := &Model{
		style:           newStyles(0, 0),
		displayedScreen: homeScreen,
		progress:        progress.New(progress.WithDefaultScaledGradient()),
	}

	if fileInfo.IsDir() {
		collection, err = generateMusicCollection(path, &model.displayedTextWidth)
		if err != nil {
			return nil, err
		}

	} else {
		model.displayedScreen = musicScreen
		newMusic := &tuiMusic{
			Music:              &music.Music{FilePath: path},
			displayedTextWidth: &model.displayedTextWidth,
		}
		err := newMusic.PopulateMusicMeta()
		if err != nil {
			return nil, err
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

	switch m.displayedScreen {
	case homeScreen:
		return m.updateHomeScreen(msg)
	case musicScreen:
		return m.updateMusicScreen(msg)
	}
	return m, nil
}

func (m *Model) View() string {

	if m.style.height == 0 {
		return "Loading..."
	}

	if m.displayedScreen == homeScreen {
		return m.homePageView()
	}
	return m.musicPageView()
}

/* END Bubbletea model */

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
			newMusic := &tuiMusic{
				Music:              &music.Music{FilePath: path},
				displayedTextWidth: textWidth,
			}
			err := newMusic.PopulateMusicMeta()
			if err != nil {
				setErr(err)
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
