package screenshot

import (
	"os"
	"path/filepath"
)

func DefaultPicsDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "."
	}
	return filepath.Join(home,
		"Pictures",
		"Frontier Developments",
		"Elite Dangerous",
	)
}
