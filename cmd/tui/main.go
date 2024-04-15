package main

import (
	"bytes"
	"cmp"
	"errors"
	"flag"
	"image"
	"log"
	"os"
	"path/filepath"
	"slices"
	"sync"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dhowden/tag"
	"github.com/qeesung/image2ascii/convert"
)

/* Main model */

type Model struct {
	path      string
	musicInfo *MusicInfoModel
	list      list.Model
	err       error // will be useful later on
}

type MusicInfoModel struct {
	width  int
	height int
	// music      *music
	data       string
	mainStyle  lipgloss.Style
	titleStyle lipgloss.Style
	otherStyle lipgloss.Style
}

func newModel(path string) *Model {
	return &Model{path: path, musicInfo: &MusicInfoModel{}}
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
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.initList(m.path, msg.Width/2, msg.Height)
		m.musicInfo = createMusicInfo(msg.Width, msg.Height)
		music := m.list.SelectedItem().(*music)
		m.musicInfo.generate(music)
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "w", "down", "s":
			music := m.list.SelectedItem().(*music)
			m.musicInfo.generate(music)
		}
	}
	return m, cmd
}

func createMusicInfo(width, height int) *MusicInfoModel {
	mainStyle := lipgloss.NewStyle().
		Width(width / 2).
		Height(height).
		Align(lipgloss.Center).
		PaddingTop(height / 10)
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA"))
		// Background(lipgloss.Color("#7D56F4"))
		// Width(22)
	model := &MusicInfoModel{
		mainStyle:  mainStyle,
		titleStyle: titleStyle,
		height:     height,
		width:      width,
	}
	return model

}

func (m *MusicInfoModel) generate(music *music) {
	f, err := os.Open(music.FilePath)
	if err != nil {
		music.err = err
		return
	}
	defer f.Close()
	fMeta, err := tag.ReadFrom(f)
	if err != nil {
		if errors.Is(err, tag.ErrNoTagsFound) {
			// TODO: add default image here
			// add default image here
		}
		music.err = err
		return
	}
	img := "no image"
	temp := fMeta.Picture()
	if temp != nil {
		pic, _, _ := image.Decode(bytes.NewReader(fMeta.Picture().Data))
		// TODO: remove this commented out code
		// picF, _ := os.Create(fmt.Sprintf("text.%s", fMeta.Picture().Ext))
		// fmt.Println(fMeta.Picture().MIMEType)
		// switch fMeta.Picture().MIMEType {
		// case "image/jpeg":
		// 	jpeg.Encode(picF, pic, &jpeg.Options{})
		// case "image/png":
		// 	png.Encode(picF, pic)

		// }
		// fmt.Println(m.height, m.width)
		img = convert.NewImageConverter().Image2ASCIIString(pic, &convert.Options{Colored: true, FixedHeight: m.height / 2})
	}
	m.data = m.mainStyle.Render(img, m.titleStyle.Render(music.Name), m.titleStyle.Render(music.Artist))
}

func (m *Model) View() string {
	return lipgloss.JoinHorizontal(lipgloss.Left, m.list.View(), m.musicInfo.data)
	// return m.list.View()
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
