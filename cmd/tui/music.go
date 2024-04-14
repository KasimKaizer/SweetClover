package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

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

func isMusic(file string) bool {
	switch filepath.Ext(file) {
	case ".wav", ".mp3", ".ogg", ".flac":
		return true
	}
	return false
}
