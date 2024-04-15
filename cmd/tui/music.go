package main

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/dhowden/tag"
	"github.com/qeesung/image2ascii/convert"
)

/* music Model */
type music struct {
	FilePath string
	Name     string
	Artist   string
	err      error
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
			m.Artist = _defaultName
			return nil
		}
		return err
	}
	m.Name = fMeta.Title()
	m.Artist = fMeta.Artist()
	if m.Name == "" {
		name := filepath.Base(f.Name())
		m.Name = strings.TrimSuffix(name, filepath.Ext(name))
	}
	return nil
}

func isMusic(file string) bool {
	switch filepath.Ext(file) {
	case ".wav", ".mp3", ".ogg", ".flac", ".m4a":
		return true
	}
	return false
}

type MusicInfoModel struct {
	name        string
	artist      string
	album       string
	releaseYear int
	img         string
	style       *styles
}

func createMusicInfo(width, height int) *MusicInfoModel {
	model := &MusicInfoModel{
		style:       newStyles(width, height),
		img:         _defaultImg,
		name:        _defaultName,
		artist:      _defaultName,
		album:       _defaultName,
		releaseYear: _defaultYear,
	}
	return model
}

func (m *MusicInfoModel) getMusicMeta(music *music) {
	f, err := os.Open(music.FilePath)
	if err != nil {
		music.err = err
		return
	}
	defer f.Close()
	fMeta, err := tag.ReadFrom(f)
	if err != nil {
		if errors.Is(err, tag.ErrNoTagsFound) {
			m.name, m.artist, m.album = music.Name, _defaultName, _defaultName
			m.releaseYear = _defaultYear
			return
		}
		music.err = err
		return
	}
	m.name, m.artist, m.album = fMeta.Title(), fMeta.Artist(), fMeta.Album()
	m.releaseYear = fMeta.Year()
	if m.name == "" {
		m.name, m.artist, m.album = music.Name, _defaultName, _defaultName
		m.releaseYear = _defaultYear
	}
	temp := fMeta.Picture()
	if temp == nil {
		m.img = _defaultImg
		return
	}
	pic, _, err := image.Decode(bytes.NewReader(fMeta.Picture().Data))
	if err != nil {
		music.err = err
		return
	}
	m.img = convert.NewImageConverter().Image2ASCIIString(pic,
		&convert.Options{
			Colored:     true,
			FitScreen:   true,
			FixedHeight: int(float64(m.style.height) * 0.6),
			FixedWidth:  int(float64(m.style.width) * 0.35),
		})
}

func (m *MusicInfoModel) format() string {
	return lipgloss.NewStyle().
		Align(lipgloss.Left, lipgloss.Center).
		PaddingTop(int(float64(m.style.height)*0.05)).
		Render(
			lipgloss.Place(m.style.width/2,
				int(float64(m.style.height)*0.7),
				lipgloss.Center, lipgloss.Center,
				m.style.imageStyle.Render(m.img),
			),
			m.style.titleStyle.Render(
				fmt.Sprintf(
					_infoBoxTmpl,
					m.style.textStyle.Render(truncate(m.name, m.style.width/3)),
					m.style.textStyle.Render(truncate(m.album, m.style.width/3)),
					m.style.textStyle.Render(truncate(m.artist, m.style.width/3)),
					m.style.textStyle.Render(strconv.Itoa(m.releaseYear)),
				)),
		)
}
