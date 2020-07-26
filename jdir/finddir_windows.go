package jdir

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
)

var relJournalPath = []string{
	"",
	"Saved Games",
	"Frontier Developments",
	"Elite Dangerous",
}

func FindJournalDir() (dir string, err error) {
	usr, err := user.Current()
	if err != nil {
		return ".", err
	}
	relJournalPath[0] = usr.HomeDir
	dir = filepath.Join(relJournalPath...)
	if stat, err := os.Stat(dir); err != nil {
		return "", err
	} else if !stat.IsDir() {
		return "", fmt.Errorf("'%s' is not a directory", dir)
	}
	return dir, nil
}
