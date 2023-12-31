package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"

	"github.com/CmdrVasquess/watched"
)

//go:generate versioner -pkg main -bno build_no VERSION version.go

const configName = "edeh.json"

var (
	fJDir           string
	fPluginPath     string
	fData           string
	fPinOff, fPinOn string
	fNet            string
	fManifests      = "plugin.json"
	fPinQLen        = 64
	fTCPQLen        = 64

	config struct {
		Version     string
		LastJournal string
		LastEvent   int
	}
	distro distributor
)

func readConfig() error {
	cfgFile := filepath.Join(fData, configName)
	log.Info("read `config`", `config`, cfgFile)
	rd, err := os.Open(cfgFile)
	switch {
	case os.IsNotExist(err):
		log.Warn("1st start, `config` not exists", `config`, cfgFile)
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
	log.Info("write `config`", `config`, cfgFile)
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
	res, err := watched.FindJournalDir()
	if err != nil {
		logFatal(err.Error())
	}
	return res
}

func mustFindDataDir() string {
	res, err := findDataDir()
	if err != nil {
		logFatal(err.Error())
	}
	return res
}

func flags() {
	flag.StringVar(&fJDir, "j", "",
		"Manually set the directory with ED's journal files")
	flag.StringVar(&fPluginPath, "p", "./plugins",
		"Set plugin path")
	flag.StringVar(&fData, "d", mustFindDataDir(), "Directory where data is stored")
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
	logFlag := flag.String("log", "", "Set logging <level>[-|+[fps]]")
	flag.Parse()
	if err := logCfg.ParseFlag(*logFlag); err != nil {
		logFatal(err.Error())
	}
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
	log.Info(fmt.Sprintf("edeh v%d.%d.%d-%s+%d", Major, Minor, Patch, Quality, BuildNo))
	flags()
	if err := readConfig(); err != nil {
		logFatal(err.Error())
	}
	// TODO check config version
	config.Version = fmt.Sprintf("%d.%d.%d-%s+%d",
		Major, Minor, Patch, Quality, BuildNo)
	opts := &watched.JournalOptions{
		SerialIndependent: []string{
			"Fileheader",
			"Commander",
			"Shutdown",
		},
	}
	watchED := watched.NewJournal(fJDir, &distro, opts)
	if fNet != "" {
		err := distro.load(fNet)
		if err != nil {
			log.Error(err.Error())
		}
		for i := range distro.TCP {
			go distro.TCP[i].runLoop(&distro)
		}
	}
	loadPlugins(
		fPluginPath,
		strings.Split(fManifests, string(filepath.ListSeparator)),
	)
	go watchED.Start()
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt)
	<-sigs
	log.Info("shutting down…")
	watchED.Stop()
	if err := writeConfig(); err != nil {
		log.Error(err.Error())
	}
	distro.Close()
	distro.waitClose.Wait()
	log.Info("bye!")
}
