package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"

	"github.com/CmdrVasquess/watched"
	"github.com/CmdrVasquess/watched/jdir"
)

//go:generate versioner -pkg main -bno build_no VERSION version.go

const configName = "edeh.json"

var (
	fJDir        string
	fWatchLatest bool
	fPluginPath  string
	fData        string

	config struct {
		Version string
		LastSer int64
	}
	disp watched.RecvToSrc
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

func journalEvent(ser int64, revt watched.RawEvent) {
	log.Infof("journal event: %s", string(revt))
	if ser <= config.LastSer {
		return
	} else {
		config.LastSer = ser
	}
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "%d ", ser)
	buf.Write(revt)
	buf.WriteByte('\n')
	for _, pin := range plugins {
		if err := pin.sendJournal(buf.Bytes()); err != nil {
			log.Errore(err)
		}
	}
}

func statusEvent(typ watched.StatusType, revt watched.RawEvent) {
	log.Infof("%s: %s", typ, string(revt))
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
	flag.StringVar(&fJDir, "j", "",
		"Manually set the directory with ED's journal files")
	flag.BoolVar(&fWatchLatest, "watch-latest", true,
		"Start with watching latest journal file")
	flag.StringVar(&fPluginPath, "p", "./plugins",
		"Set plugin path")
	flag.StringVar(&fData, "d", mustFindDataDir(), "Directory where data is stored")
	flag.Parse()
	if fJDir == "" {
		fJDir = mustJournalDir()
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
	watchED := jdir.NewEvents(fJDir, &disp, nil)
	loadPlugins(fPluginPath)
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
	log.Infos("bye!")
}
