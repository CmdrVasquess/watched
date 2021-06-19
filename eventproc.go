package watched

import (
	"errors"
	"time"
)

type RawEvent []byte

func (re RawEvent) PeekTime() (time.Time, error) {
	return PeekTime(re)
}

func (re RawEvent) PeekEvent() (string, error) {
	return PeekEvent(re)
}

func (re RawEvent) Peek() (time.Time, string, error) {
	return Peek(re)
}

type StatusType int

func (st StatusType) String() string { return statNames[st] }

func ParseStatusType(s string) StatusType {
	for i, n := range statNames[1:] {
		if n == s {
			return StatusType(i + 1)
		}
	}
	return 0
}

const (
	StatCargo StatusType = iota + 1
	StatMarket
	StatModules
	StatNavRoute
	StatOutfit
	StatShipyard
	StatStatus
)

const (
	StatCargoName    = "Cargo"
	StatMarketName   = "Market"
	StatModulesName  = "ModuleInfo"
	StatNavRouteName = "NavRoute"
	StatOutfitName   = "Outfitting"
	StatShipyardName = "Shipyard"
	StatStatusName   = "Status"
)

type JEventID = int64

const StartNow JEventID = -1

type JounalEvent struct {
	Serial JEventID
	Event  RawEvent
}

type StatusEvent struct {
	Type  StatusType
	Event RawEvent
}

type EventSrc struct {
	Journal <-chan JounalEvent
	Status  <-chan StatusEvent
}

type EventRecv interface {
	OnJournalEvent(e JounalEvent) error
	OnStatusEvent(e StatusEvent) error
	Close() error
}

var statNames = []string{
	"<non-status>",
	StatCargoName, StatMarketName, StatModulesName, StatNavRouteName,
	StatOutfitName, StatShipyardName, StatStatusName,
}

const (
	jeSequenceBits = 10
	jeSequenceMask = (1 << jeSequenceBits) - 1
)

type JEIDCounter struct {
	lastUnix int64
	seq      int64
}

func (idc *JEIDCounter) Count(t time.Time) (JEventID, error) {
	return idc.CountUnix(t.Unix())
}

func (idc *JEIDCounter) CountUnix(tu int64) (JEventID, error) {
	tu <<= jeSequenceBits
	switch {
	case tu < idc.lastUnix:
		return 0, errors.New("JEventID timestamp out of sequence")
	case tu > idc.lastUnix:
		idc.lastUnix = tu
		idc.seq = 0
		return tu, nil
	}
	idc.seq++
	if idc.seq|jeSequenceMask != jeSequenceMask {
		return 0, errors.New("JEeventID sequence overflow")
	}
	return tu | idc.seq, nil
}

func (idc *JEIDCounter) SetLast(jeid JEventID) {
	idc.lastUnix = jeid & ^jeSequenceMask
	idc.seq = jeid & jeSequenceMask
}

func (idc *JEIDCounter) Last() JEventID {
	return idc.lastUnix | idc.seq
}
