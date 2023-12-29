package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"git.fractalqb.de/fractalqb/qblog"
	plugin "github.com/CmdrVasquess/watched/edeh/plugin"
	"github.com/CmdrVasquess/watched/examples/speak"
	"gopkg.in/yaml.v3"
)

var (
	log = qblog.New(&qblog.DefaultConfig).WithGroup("speak")

	speaker speak.Speaker
)

func usage() {
	w := flag.CommandLine.Output()
	fmt.Fprintf(w, "Usage: %s [flags] <config>...\n", os.Args[0])
	flag.PrintDefaults()
}

func main() {
	flag.Usage = usage
	speak.SetLog(log.Logger)
	fVerb := flag.Bool("v", false, "Verbose output")
	flag.Parse()
	log.Info("start edeh plugin 'speak'")
	for _, arg := range flag.Args() {
		switch filepath.Ext(arg) {
		case ".json":
			readCfgJSON(arg)
		case ".yaml", ".yml":
			readCfgYAML(arg)
		}
	}
	speaker.Verbose = speaker.Verbose || *fVerb
	defer speaker.Close()
	plugin.RunRecv(&speaker, nil, slog.Default())
}

func logFatal(msg string, args ...any) {
	log.Error(msg, args...)
	os.Exit(1)
}

func readCfgJSON(name string) {
	log.Info("read `config`", `config`, name)
	rd, err := os.Open(name)
	if err != nil {
		logFatal(err.Error())
	}
	defer rd.Close()
	dec := json.NewDecoder(rd)
	if err = dec.Decode(&speaker); err != nil {
		logFatal(err.Error())
	}
	if err = speaker.Configure(); err != nil {
		logFatal(err.Error())
	}
}

func readCfgYAML(name string) {
	log.Info("read `config`", `config`, name)
	rd, err := os.Open(name)
	if err != nil {
		logFatal(err.Error())
	}
	defer rd.Close()
	dec := yaml.NewDecoder(rd)
	if err = dec.Decode(&speaker); err != nil {
		logFatal(err.Error())
	}
	if err = speaker.Configure(); err != nil {
		logFatal(err.Error())
	}
}
