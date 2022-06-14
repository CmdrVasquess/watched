package main

import (
	"git.fractalqb.de/fractalqb/c4hgol"
	"git.fractalqb.de/fractalqb/qblog"
	"github.com/CmdrVasquess/watched"
)

var (
	log    = qblog.New("edeh")
	logCfg = c4hgol.NewLogGroup(log, "", watched.LogCfg())
)
