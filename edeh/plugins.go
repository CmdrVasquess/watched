package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/CmdrVasquess/watched"

	"git.fractalqb.de/fractalqb/sllm"
)

const pluginManifest = "plugin.json"

type bwList []string

func (l bwList) Len() int { return len(l) }

func (l bwList) Index(e string) int {
	for i, f := range l {
		if e == f {
			return i
		}
	}
	return -1
}

func (l bwList) Blacklisted(e string) bool {
	switch {
	case l == nil:
		return false
	case len(l) == 0:
		return true
	}
	return l.Index(e) >= 0
}

func (l bwList) Whitelisted(e string) bool {
	return l.Index(e) >= 0
}

type plugin struct {
	Name    string
	Off     bool
	Run     string
	Args    []string `json:",omitempty"`
	Wdir    string   `json:",omitempty"`
	Stdout  bool     `json:",omitempty"`
	Stderr  bool     `json:",omitempty"`
	Journal struct {
		Blacklist bwList
		Whitelist bwList
	}
	rootDir string
	cmd     *exec.Cmd
	pipe    io.WriteCloser
}

func (pin *plugin) sendJournal(line watched.RawEvent) error {
	event, err := line.PeekEvent()
	if err != nil {
		return err
	}
	if pin.Journal.Blacklist.Blacklisted(event) &&
		!pin.Journal.Whitelist.Whitelisted(event) {
		return nil
	}
	if _, err := pin.pipe.Write(line); err != nil {
		log.Warna("sending journal `event` `to`: `err`", event, pin.Name, err)
	}
	return nil
}

var plugins []*plugin

func loadPlugins(path string) {
	pdirs := filepath.SplitList(path)
	for _, dir := range pdirs {
		loadPluginsDir(dir)
	}
}

func loadPluginsDir(dir string) {
	log.Infoa("search plugins in `dir`", dir)
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && info.Name() == pluginManifest {
			if err := loadPlugin(path); err != nil {
				log.Errore(err)
			}
		}
		return nil
	})
	if err != nil {
		log.Fatale(err)
	}
}

func readPluginManifest(file string) (*plugin, error) {
	rd, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer rd.Close()
	res := new(plugin)
	err = json.NewDecoder(rd).Decode(res)
	res.rootDir = filepath.Dir(file)
	return res, err
}

func checkRunPath(pin *plugin) error {
	run := filepath.Clean(filepath.Join(pin.rootDir, pin.Run))
	if !filepath.HasPrefix(run, pin.rootDir) {
		return sllm.Error("`run path` of `plugin` not in `plugin dir`",
			pin.Run,
			pin.Name,
			pin.rootDir,
		)
	}
	pin.Run = run
	return nil
}

func loadPlugin(manifest string) error {
	pin, err := readPluginManifest(manifest)
	if err != nil {
		return fmt.Errorf("load manifest '%s': %s", manifest, err)
	}
	if pin.Off {
		log.Infoa("`plugin` is switched off", pin.Name)
		return nil
	}
	log.Infoa("load `plugin` from `dir`", pin.Name, pin.rootDir)
	if err = checkRunPath(pin); err != nil {
		return err
	}
	pin.cmd = exec.Command(pin.Run, pin.Args...)
	if pin.Stdout {
		pin.cmd.Stdout = os.Stdout
	}
	if pin.Stderr {
		pin.cmd.Stderr = os.Stderr
	}
	pin.pipe, err = pin.cmd.StdinPipe()
	if err != nil {
		return err
	}
	pin.cmd.Dir = pin.Wdir
	switch {
	case len(pin.Journal.Blacklist) == 0:
		pin.Journal.Blacklist = nil
	case pin.Journal.Blacklist.Index("*") >= 0:
		pin.Journal.Blacklist = []string{}
	}
	if err = pin.cmd.Start(); err != nil {
		log.Errore(err)
	}
	plugins = append(plugins, pin)
	return nil
}
