package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"

	"git.fractalqb.de/fractalqb/c4hgol"

	"github.com/CmdrVasquess/watched"
	"github.com/CmdrVasquess/watched/jdir"
)

//go:generate versioner -pkg main -bno build_no VERSION version.go

const configName = "edeh.json"

var (
	fLog            string
	fJDir           string
	fWatchLatest    bool
	fPluginPath     string
	fData           string
	fOld            bool
	fPinOff, fPinOn string
	fNet            string
	fManifests      = "plugin.json"
	fPinQLen        = 64
	fTCPQLen        = 64

	config struct {
		Version string
		LastSer int64
	}
	distro distributor
)

func readConfig() error {
	cfgFile := filepath.Join(fData, configName)
	log.Infoa("read `config`", cfgFile)
	rd, err := os.Open(cfgFile)
	switch {
	case os.IsNotExist(err):
		log.Warna("1st start, `config` not exists", cfgFile)
		return nil
	case err != nil:
		return err
	}
	defer rd.Close()
	dec := json.NewDecoder(rd)
	return dec.Decode(&config)
}

func writeConfig() error {
	cfgFile := filepath.Join(fData, configName)
	tmpFile := cfgFile + "~"
	log.Infoa("write `config`", cfgFile)
	wr, err := os.Create(tmpFile)
	if err != nil {
		return err
	}
	defer wr.Close()
	enc := json.NewEncoder(wr)
	enc.SetIndent("", "\t")
	if err = enc.Encode(&config); err != nil {
		return err
	}
	wr.Close()
	return os.Rename(tmpFile, cfgFile)
}

func mustJournalDir() string {
	res, err := jdir.FindJournalDir()
	if err != nil {
		log.Fatale(err)
	}
	return res
}

func mustFindDataDir() string {
	res, err := findDataDir()
	if err != nil {
		log.Fatale(err)
	}
	return res
}

func flags() {
	flag.StringVar(&fLog, c4hgol.DefaultFlagLevel, "", c4hgol.LevelCfgDoc(nil))
	flag.StringVar(&fJDir, "j", "",
		"Manually set the directory with ED's journal files")
	flag.BoolVar(&fWatchLatest, "watch-latest", true,
		"Start with watching latest journal file")
	flag.StringVar(&fPluginPath, "p", "./plugin",
		"Set plugin path")
	flag.StringVar(&fData, "d", mustFindDataDir(), "Directory where data is stored")
	flag.BoolVar(&fOld, "old", false, "Accept past events")
	flag.StringVar(&fPinOff, "off", "",
		"Comma separated list of plugins to switch off")
	flag.StringVar(&fPinOn, "on", "",
		"Comma separated list of plugins to switch on")
	flag.StringVar(&fNet, "net", "", "Load net configuration file")
	flag.StringVar(&fManifests, "manifests", fManifests,
		"List of potenitial manifest file names separated by '"+
			string(filepath.ListSeparator)+`'. Use this if
you want manifest files with non-default names. 1st match will be
loaded.`)
	flag.IntVar(&fPinQLen, "pq-len", fPinQLen,
		"Length of plugin event queues")
	flag.IntVar(&fTCPQLen, "tcpq-len", fTCPQLen,
		"Default length of TCP client event queues")
	flag.Parse()
	c4hgol.SetLevel(logCfg, fLog, nil)
	if fJDir == "" {
		fJDir = mustJournalDir()
	}
	if fPinOff != "" {
		pins := strings.Split(fPinOff, ",")
		for _, pin := range pins {
			pinSwitches[pin] = false
		}
	}
	if fPinOn != "" {
		pins := strings.Split(fPinOn, ",")
		for _, pin := range pins {
			pinSwitches[pin] = true
		}
	}
}

func main() {
	log.Infof("edeh v%d.%d.%d-%s+%d", Major, Minor, Patch, Quality, BuildNo)
	flags()
	if err := readConfig(); err != nil {
		log.Fatale(err)
	}
	// TODO check config version
	config.Version = fmt.Sprintf("%d.%d.%d-%s+%d",
		Major, Minor, Patch, Quality, BuildNo)
	var latestJournal string
	if fWatchLatest {
		var err error
		latestJournal, err = jdir.NewestJournal(fJDir)
		if err != nil {
			log.Fatale(err)
		}
	}
	opts := &jdir.Options{
		SerialIndependent: []string{
			"Fileheader",
			"Commander",
			"Shutdown",
		},
	}
	if !fOld {
		opts.JSerial = watched.StartNow
	}
	watchED := jdir.NewEvents(fJDir, &distro, opts)
	if fNet != "" {
		err := distro.load(fNet)
		if err != nil {
			log.Errore(err)
		}
		for i := range distro.TCP {
			go distro.TCP[i].runLoop(&distro.reconnect)
		}
	}
	loadPlugins(
		fPluginPath,
		strings.Split(fManifests, string(filepath.ListSeparator)),
	)
	go watchED.Start(latestJournal)
	sigs := make(chan os.Signal)
	signal.Notify(sigs, os.Interrupt)
	<-sigs
	log.Infos("shutting down…")
	watchED.Stop <- watched.Stop
	<-watchED.Stop
	if err := writeConfig(); err != nil {
		log.Errore(err)
	}
	distro.waitClose.Wait()
	log.Infos("bye!")
}
