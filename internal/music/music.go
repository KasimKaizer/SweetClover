package music

import (
	"bytes"
	"errors"
	"image"
	"os"
	"path/filepath"
	"strings"

	"github.com/dhowden/tag"
	"github.com/qeesung/image2ascii/convert"
)

const (
	_defaultName = "UNKNOWN"
	_defaultImg  = "NO_IMAGE"
	_defaultYear = 2100
)

type Music struct {
	FilePath    string
	Name        string
	Artist      string
	Album       string
	ReleaseYear int
}

func NewMusic(path string) (*Music, error) {
	model := &Music{
		FilePath: path,
	}
	err := model.PopulateMusicMeta()
	return model, err
}

func (m *Music) FilterValue() string {
	return m.Name
}

func (m *Music) Title() string {
	return m.Name
}

func (m *Music) Description() string {
	return m.Artist
}

func (m *Music) PopulateMusicMeta() error {
	f, err := os.Open(m.FilePath)
	if err != nil {
		return err
	}
	defer f.Close()
	fMeta, err := tag.ReadFrom(f)
	if err != nil || fMeta.Title() == "" {
		if err != nil && !errors.Is(err, tag.ErrNoTagsFound) {
			return err
		}
		name := filepath.Base(f.Name())
		m.Name = strings.TrimSuffix(name, filepath.Ext(name))
		m.Artist, m.Album = _defaultName, _defaultName
		m.ReleaseYear = _defaultYear
		return nil
	}
	m.Name, m.Artist, m.Album = fMeta.Title(), fMeta.Artist(), fMeta.Album()
	m.ReleaseYear = fMeta.Year()
	return nil
}

func (m *Music) GetCoverArtASCII(height, width int) (string, error) {
	f, err := os.Open(m.FilePath)
	if err != nil {
		return _defaultImg, err
	}
	defer f.Close()
	fMeta, err := tag.ReadFrom(f)
	if err != nil && !errors.Is(err, tag.ErrNoTagsFound) {
		return _defaultImg, nil
	}
	if err == nil && fMeta.Picture() == nil {
		return _defaultImg, nil
	}
	pic, _, err := image.Decode(bytes.NewReader(fMeta.Picture().Data))
	if err != nil {
		return _defaultImg, nil
	}
	image := convert.NewImageConverter().Image2ASCIIString(pic,
		&convert.Options{
			Colored:     true,
			FitScreen:   true,
			FixedHeight: height,
			FixedWidth:  width,
			// FixedHeight: int(float64(m.style.height) * 0.6),
			// FixedWidth:  int(float64(m.style.width) * 0.35),
		})
	return image, nil
}

func IsMusic(file string) bool {
	switch filepath.Ext(file) {
	case ".wav", ".mp3", ".ogg", ".flac", ".m4a":
		return true
	}
	return false
}
