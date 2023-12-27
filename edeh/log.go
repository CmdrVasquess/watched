package main

import (
	"os"

	"git.fractalqb.de/fractalqb/qblog"
	"github.com/CmdrVasquess/watched"
)

var (
	logCfg = qblog.DefaultConfig.Clone().SetTimeFmt(qblog.TMillis)
	log    = qblog.New(logCfg)
)

func init() {
	watched.SetLog(log.WithGroup("watched").Logger)
}

func logFatal(msg string, args ...any) {
	log.Error(msg, args...)
	os.Exit(1)
}
