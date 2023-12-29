package speak

import (
	"log/slog"
	"os"
)

var log = slog.Default()

func SetLog(l *slog.Logger) {
	log = l
}

func logFatal(err error) {
	log.Error(err.Error())
	os.Exit(1)
}
