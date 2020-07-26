package internal

import (
	"git.fractalqb.de/fractalqb/c4hgol"
	"git.fractalqb.de/fractalqb/qbsllm"
)

var (
	JDirLog = qbsllm.New(qbsllm.Lnormal, "jdir", nil, nil)
	RootLog = qbsllm.New(qbsllm.Lnormal, "watchED", nil, nil)
	LogCfg  = c4hgol.Config(qbsllm.NewConfig(RootLog),
		qbsllm.NewConfig(JDirLog),
	)
)
