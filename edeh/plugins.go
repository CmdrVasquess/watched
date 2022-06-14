package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"git.fractalqb.de/fractalqb/sllm/v2"
	"github.com/CmdrVasquess/watched"
)

const shutdownDelay = 5 * time.Second

var pinSwitches = make(map[string]bool)

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

type BlackWhiteList struct {
	Blacklist bwList
	Whitelist bwList
}

func (bw *BlackWhiteList) Filter(s string) bool {
	if bw.Blacklist.Blacklisted(s) {
		return bw.Whitelist.Whitelisted(s)
	}
	return true
}

type plugin struct {
	// The name of the plugin, just for documentation
	Name string
	// The plugin will onyl be started if !Off
	Off bool
	// The command to exec when the plugin is stared. Run is checked to be
	// inside or below its plugin folder. If not, the plugin will be ignored.
	Run string
	// The arguments that are passed to the Run command.
	Args []string `json:",omitempty"`
	// If Stdout is true, the stdout of the plugin will be connected to
	// EDEH's stdout. Otherwise it goes to /dev/null.
	Stdout bool `json:",omitempty"`
	// If Stderr is true, the stderr of the plugin will be connected to
	// EDEH's stderr. Otherwise it goes to /dev/null.
	Stderr  bool `json:",omitempty"`
	Journal BlackWhiteList
	Status  BlackWhiteList

	rootDir string
	cmd     *exec.Cmd
	pipe    io.WriteCloser
	jes     chan *jEvent
	ses     chan *sEvent
}

type jEvent struct {
	watched.JounalEvent
	evt string
	msg []byte
}

type sEvent struct {
	watched.StatusEvent
	msg []byte
}

func (pin *plugin) start(closed *sync.WaitGroup) {
	closed.Add(1)
	defer closed.Done()
	log.Infov("running receive loop of `plugin`", pin.Name)
	count := 0
	if pin.jes != nil {
		count++
	}
	if pin.ses != nil {
		count++
	}
	for count > 0 {
		select {
		case e, ok := <-pin.jes:
			if ok {
				if err := pin.sendJournal(e); err != nil {
					log.Warnv("sending journal `event` `to`: `err`",
						e.evt,
						pin.Name,
						err)
				}
			} else {
				count--
			}
		case e, ok := <-pin.ses:
			if ok {
				if err := pin.sendStatus(e); err != nil {
					log.Warnv("sending status `event` `to`: `err`",
						e.Type,
						pin.Name,
						err)
				}
			} else {
				count--
			}
		}
	}
	log.Debugv("leave receive loop of `plugin`, shutdown…", pin.Name)
	if err := pin.pipe.Close(); err != nil {
		log.Errorv("closing pipe to `plugin`: `err`", pin.Name, err)
		pin.cmd.Process.Kill()
		log.Warnv("killed `plugin`", pin.Name)
	} else {
		t := time.AfterFunc(shutdownDelay, func() {
			log.Warnv("`shutdown time` of `plugin` exceeded, kill", shutdownDelay, pin.Name)
			pin.cmd.Process.Kill()
		})
		pin.cmd.Wait()
		t.Stop()
		log.Infov("shutdown of `plugin` done", pin.Name)
	}
}

func (pin *plugin) sendJournal(je *jEvent) error {
	if pin.Journal.Filter(je.evt) {
		_, err := pin.pipe.Write(je.msg)
		return err
	}
	return nil
}

func (pin *plugin) sendStatus(se *sEvent) error {
	if pin.Status.Filter(se.Type.String()) {
		_, err := pin.pipe.Write(se.msg)
		return err
	}
	return nil
}

func loadPlugins(path string, manifests []string) {
	pdirs := filepath.SplitList(path)
	for _, dir := range pdirs {
		loadPluginsDir(dir, manifests)
	}
}

func loadPluginsDir(dir string, manifests []string) {
	for _, m := range manifests {
		mf := filepath.Join(dir, m)
		if _, err := os.Stat(mf); err == nil {
			if err = loadPlugin(mf); err != nil {
				log.Errore(err)
			}
			return
		}
	}
	ls, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Errore(err)
		return
	}
	for _, l := range ls {
		if !l.IsDir() {
			continue
		}
		loadPluginsDir(filepath.Join(dir, l.Name()), manifests)
	}
}

// func loadPluginsDir(dir string) {
// 	log.Infov("search plugins in `dir`", dir)
// 	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
// 		if err != nil {
// 			return err
// 		}
// 		if !info.IsDir() && info.Name() == pluginManifest {
// 			if err := loadPlugin(path); err != nil {
// 				log.Errore(err)
// 			}
// 		}
// 		return nil
// 	})
// 	if err != nil {
// 		log.Fatale(err)
// 	}
//}

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
	if run, ok := pinSwitches[pin.Name]; ok {
		if !run {
			log.Infov("`plugin` is switched off", pin.Name)
			return nil
		}
	} else if pin.Off {
		log.Infov("`plugin` is switched off", pin.Name)
		return nil
	}
	log.Infov("load `plugin` from `dir`", pin.Name, pin.rootDir)
	if err = checkRunPath(pin); err != nil {
		return err
	}
	exe, err := filepath.Abs(pin.Run)
	if err != nil {
		return err
	}
	pin.cmd = exec.Command(exe, pin.Args...)
	pin.cmd.Dir = pin.rootDir
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
	switch {
	case len(pin.Journal.Blacklist) == 0:
		pin.Journal.Blacklist = nil
	case pin.Journal.Blacklist.Index("*") >= 0:
		pin.Journal.Blacklist = []string{}
	}
	if err = pin.cmd.Start(); err != nil {
		return err
	}
	distro.addPlugin(pin)
	go pin.start(&distro.waitClose)
	return nil
}
