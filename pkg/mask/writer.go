package mask

import (
	"io"
	"strings"

	"github.com/mercari/tfnotify/v1/pkg/config"
)

const (
	typeEqual  = "equal"
	typeRegexp = "regexp"
)

type Writer struct {
	patterns []*config.Mask
	w        io.Writer
}

func NewWriter(w io.Writer, patterns []*config.Mask) *Writer {
	return &Writer{
		w:        w,
		patterns: patterns,
	}
}

func (w *Writer) Write(p []byte) (int, error) {
	a := p
	for _, pattern := range w.patterns {
		switch pattern.Type {
		case typeEqual:
			a = []byte(strings.ReplaceAll(string(a), pattern.Value, "***"))
		case typeRegexp:
			a = pattern.Regexp.ReplaceAll(a, []byte("***"))
		}
	}
	_, err := w.w.Write(a)
	return len(p), err
}

func Mask(s string, patterns []*config.Mask) string {
	a := s
	for _, pattern := range patterns {
		switch pattern.Type {
		case typeEqual:
			a = strings.ReplaceAll(a, pattern.Value, "***")
		case typeRegexp:
			a = pattern.Regexp.ReplaceAllString(a, "***")
		}
	}
	return a
}
