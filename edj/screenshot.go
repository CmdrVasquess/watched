package edj

import (
	"path/filepath"
	"strings"
)

const ScreenshotTag = "Screenshot"

type Screenshot struct {
	Event
	Filename      string
	Width, Height int
	System        string
	Body          string
}

func (s *Screenshot) FilenameToOS() string {
	return strings.ReplaceAll(s.Filename, "\\", string(filepath.Separator))
}
