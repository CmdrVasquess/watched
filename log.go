package watched

import (
	"log/slog"

	"git.fractalqb.de/fractalqb/qblog"
)

var log = qblog.New(&qblog.DefaultConfig)

func SetLog(l *slog.Logger) { log = &qblog.Logger{l} }
