package internal

import (
	"git.fractalqb.de/fractalqb/c4hgol"
	"git.fractalqb.de/fractalqb/qblog"
)

var (
	JDirLog = qblog.New("jdir")
	RootLog = qblog.New("watchED")
	LogCfg  = c4hgol.NewLogGroup(RootLog, "", JDirLog)
)
