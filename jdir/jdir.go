package jdir

import (
	"os"
	str "strings"
	"time"

	"github.com/CmdrVasquess/watched"
	"github.com/CmdrVasquess/watched/internal"
)

var log = internal.JDirLog

var statsFiles = map[string]watched.StatusType{
	"Cargo.json":       watched.StatCargo,
	"Market.json":      watched.StatMarket,
	"ModulesInfo.json": watched.StatModules,
	"NavRoute.json":    watched.StatNavRoute,
	"Outfitting.json":  watched.StatOutfit,
	"Shipyard.json":    watched.StatShipyard,
	"Status.json":      watched.StatStatus,
}

func IsJournalFile(name string) bool {
	return str.HasPrefix(name, "Journal.") &&
		str.HasSuffix(name, ".log")
}

func NewestJournal(inDir string) (res string, err error) {
	dir, err := os.Open(inDir)
	if err != nil {
		return "", err
	}
	defer dir.Close()
	var maxTime time.Time
	infos, err := dir.Readdir(1)
	for len(infos) > 0 && err == nil {
		info := infos[0]
		if IsJournalFile(info.Name()) && (info.ModTime().After(maxTime) || len(res) == 0) {
			res = info.Name()
			maxTime = info.ModTime()
		}
		infos, err = dir.Readdir(1)
	}
	return res, nil
}
