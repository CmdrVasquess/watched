// +build !windows

package main

import (
	"os"
	"os/user"
	"path/filepath"
)

func findDataDir() (dir string, err error) {
	user, err := user.Current()
	if err != nil {
		return ".", nil
	}
	dir = filepath.Join(user.HomeDir, ".local", "share")
	if _, err = os.Stat(dir); os.IsNotExist(err) {
		return ".", nil
	}
	dir = filepath.Join(dir, "ED-event_hub")
	os.MkdirAll(dir, 0777)
	return dir, nil
}
