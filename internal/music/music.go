// Package music contains various tools to work with music files, that includes parsing then,
// getting metadata and playing these music files.
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
	"github.com/gopxl/beep/flac"
	"github.com/gopxl/beep/mp3"
	"github.com/gopxl/beep/speaker"
	"github.com/gopxl/beep/vorbis"
	"github.com/gopxl/beep/wav"
	"github.com/qeesung/image2ascii/convert"
)

// defaults for metadata.
const (
	_defaultName = "UNKNOWN"
	_defaultImg  = "NO_IMAGE"
	_defaultYear = 2100
)

// Music struct represents aa music file with its metadata.
type Music struct {
	filePath    string
	Name        string
	Artist      string
	Album       string
	ReleaseYear int
	Format      string
}

// NewMusic takes a path to a music file, and returns a music type withMetadata representing that file.
func NewMusic(path string) (*Music, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	format := filepath.Ext(path)

	fMeta, err := tag.ReadFrom(f)
	if err != nil || fMeta.Title() == "" {
		if err != nil && !errors.Is(err, tag.ErrNoTagsFound) {
			return nil, err
		}
		base := filepath.Base(f.Name())
		name := strings.TrimSuffix(base, filepath.Ext(base))
		model := &Music{
			filePath:    path,
			Name:        name,
			Artist:      _defaultName,
			Album:       _defaultName,
			ReleaseYear: _defaultYear,
			Format:      format,
		}
		return model, nil
	}
	model := &Music{
		filePath:    path,
		Name:        fMeta.Title(),
		Artist:      fMeta.Artist(),
		Album:       fMeta.Album(),
		ReleaseYear: fMeta.Year(),
		Format:      format,
	}
	return model, nil
}

// GetCoverArtASCII method returns the cover for the music with the passed size, if present.
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
		return _defaultImg, err
	}
	asciiString := convert.NewImageConverter().Image2ASCIIString(pic,
		&convert.Options{
			Colored:         true,
			FitScreen:       true,
			FixedHeight:     height,
			FixedWidth:      width,
			Reversed:        false,
			Ratio:           1,
			StretchedScreen: false,
		})
	return asciiString, nil
}

// Controller type controls playing music, and also represented its progress.
type Controller struct {
	speakerInitialized bool
	ctrl               *beep.Ctrl
	format             beep.Format
	Done               chan struct{}
}

func NewController() *Controller {
	var format beep.Format

	done := make(chan struct{})
	return &Controller{
		Done:               done,
		ctrl:               new(beep.Ctrl),
		speakerInitialized: false,
		format:             format,
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
	case ".flac":
		streamer, format, err = flac.Decode(f)
	case ".wav":
		streamer, format, err = wav.Decode(f)
	case ".ogg":
		streamer, format, err = vorbis.Decode(f)
	default:
		return errors.New("music.Controller.Play: unsupported format")
	}
	if err != nil {
		_ = f.Close()
		return err
	}

	if !c.speakerInitialized {
		err = speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10)) //nolint:gomnd // fineTuning
		if err != nil {
			return err
		}
		c.speakerInitialized = true
	}

	speaker.Clear()

	ctrl := &beep.Ctrl{Streamer: streamer, Paused: false}
	speaker.Play(beep.Seq(ctrl, beep.Callback(func() {
		c.Done <- struct{}{}
	})))
	c.format = format
	c.ctrl = ctrl

	return nil
}

func (c *Controller) Progress() (float64, error) {
	speaker.Lock()
	defer speaker.Unlock()
	streamer, err := retrieveStreamer(c.ctrl)
	if err != nil {
		return 0, err
	}
	pos := float64(streamer.Position()) / float64(streamer.Len())
	return pos, nil
}

func (c *Controller) SeekSeconds(step int) error {
	speaker.Lock()
	defer speaker.Unlock()
	streamer, err := retrieveStreamer(c.ctrl)
	if err != nil {
		return err
	}
	seekTo := streamer.Position() + (step * c.format.SampleRate.N(time.Second))
	if seekTo < 0 {
		seekTo = 0
	}
	if seekTo > streamer.Len() {
		seekTo = streamer.Len()
	}
	wasPaused := c.ctrl.Paused
	c.ctrl.Paused = true
	_ = streamer.Seek(seekTo)
	c.ctrl.Paused = wasPaused
	return nil
}

func (c *Controller) PauseResume() {
	speaker.Lock()
	defer speaker.Unlock()
	c.ctrl.Paused = !c.ctrl.Paused
}

func IsMusic(file string) bool {
	switch filepath.Ext(file) {
	case ".wav", ".mp3", ".ogg", ".flac", ".m4a":
		return true
	}
	return false
}

func retrieveStreamer(ctrl *beep.Ctrl) (beep.StreamSeeker, error) {
	if ctrl == nil {
		return nil, errors.New("music.Progress: streamer doesn't exist")
	}
	streamer, ok := ctrl.Streamer.(beep.StreamSeeker)
	if !ok {
		return nil, errors.New("music.Progress: streamer is not a streamSeeker")
	}
	return streamer, nil
}
