package watched

import (
	"fmt"
	"os"
	"path/filepath"
)

var relJournalPath = []string{
	"",
	"Saved Games",
	"Frontier Developments",
	"Elite Dangerous",
}

func FindJournalDir() (string, error) {
	var err error
	relJournalPath[0], err = os.UserHomeDir()
	if err != nil {
		return "", err
	}
	res := filepath.Join(relJournalPath...)
	if stat, err := os.Stat(res); err != nil {
		return "", err
	} else if !stat.IsDir() {
		return "", fmt.Errorf("%s is not a directory", res)
	}
	return res, nil
}
