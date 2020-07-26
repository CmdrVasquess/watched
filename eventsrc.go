package watched

import (
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

type JounalEvent struct {
	Serial JEventID
	Event  RawEvent
}

type StatusEvent struct {
	Type  StatusType
	Event RawEvent
}

type EDEvents struct {
	Journal <-chan JounalEvent
	Status  <-chan StatusEvent
}

var statNames = []string{
	"<non-status>",
	StatCargoName, StatMarketName, StatModulesName, StatNavRouteName,
	StatOutfitName, StatShipyardName, StatStatusName,
}
