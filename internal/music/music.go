package music

import (
	"bytes"
	"errors"
	"image"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/dhowden/tag"
	"github.com/gopxl/beep"
	"github.com/gopxl/beep/mp3"
	"github.com/gopxl/beep/speaker"
	"github.com/qeesung/image2ascii/convert"
)

type playerState int

const (
	stopped playerState = iota
	paused
	playing
)

const (
	_defaultName = "UNKNOWN"
	_defaultImg  = "NO_IMAGE"
	_defaultYear = 2100
)

type Music struct {
	filePath    string
	Name        string
	Artist      string
	Album       string
	ReleaseYear int
	Format      string
}

func NewMusic(path string) (*Music, error) {
	model := &Music{
		filePath: path,
	}
	err := model.PopulateMusicMeta()
	return model, err
}

func (m *Music) PopulateMusicMeta() error {
	f, err := os.Open(m.filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	m.Format = filepath.Ext(m.filePath)

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
	f, err := os.Open(m.filePath)
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
		})
	return image, nil
}

/*
func main() {
	f, err := os.Open("/Users/kaizersuterwala/projects_go/sweet_clover/test_files/FLETCHER/FLETCHER - Becky's So Hot (2022) [MP3] [UnknownB-UnknownkHz]/01. FLETCHER - Becky's So Hot.mp3")
	if err != nil {
		log.Fatal(err)
	}

	streamer, format, err := mp3.Decode(f)
	if err != nil {
		log.Fatal(err)
	}
	defer streamer.Close()

	speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))

	done := make(chan bool)
	speaker.Play(beep.Seq(streamer, beep.Callback(func() {
		done <- true
	})))

	<-done
}
*/

type Controller struct {
	speakerInitialized bool
	streamer           *beep.Ctrl
	Done               chan struct{}
}

func NewController() *Controller {
	done := make(chan struct{})
	return &Controller{
		Done: done,
	}
}

func (c *Controller) Play(m *Music) error {
	f, err := os.Open(m.filePath)
	if err != nil {
		return err
	}
	var streamer beep.StreamSeeker
	var format beep.Format
	switch m.Format {
	case ".mp3":
		streamer, format, err = mp3.Decode(f)
	}
	if err != nil {
		f.Close()
		return err

	}
	if !c.speakerInitialized {
		speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
		c.speakerInitialized = true
	}
	speaker.Clear()
	ctrl := &beep.Ctrl{Streamer: streamer, Paused: false}
	speaker.Play(beep.Seq(ctrl, beep.Callback(func() {
		c.Done <- struct{}{}
	})))
	c.streamer = ctrl
	return nil
}

func (c *Controller) Pause() {
	speaker.Lock()
	defer speaker.Unlock()
	c.streamer.Paused = !c.streamer.Paused
}

func IsMusic(file string) bool {
	switch filepath.Ext(file) {
	case ".wav", ".mp3", ".ogg", ".flac", ".m4a":
		return true
	}
	return false
}
