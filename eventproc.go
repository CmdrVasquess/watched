package watched

import (
	"bytes"
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
	StatBackpack StatusType = iota + 1
	StatCargo
	StatFCMats
	StatMarket
	StatModules
	StatNavRoute
	StatOutfit
	StatLocker
	StatShipyard
	StatStatus

	EndStatusType
)

const (
	StatBackpackName = "Backpack"
	StatCargoName    = "Cargo"
	StatFCMatsName   = "FCMaterials"
	StatMarketName   = "Market"
	StatModulesName  = "Modules"
	StatNavRouteName = "NavRoute"
	StatOutfitName   = "Outfitting"
	StatLockerName   = "ShipLocker"
	StatShipyardName = "Shipyard"
	StatStatusName   = "Status"
)

type JounalEvent struct {
	File    string
	EventNo int
	Event   RawEvent
}

func (e *JounalEvent) Clone() JounalEvent {
	return JounalEvent{
		File:    e.File,
		EventNo: e.EventNo,
		Event:   bytes.Clone(e.Event),
	}
}

type StatusEvent struct {
	Type  StatusType
	Event RawEvent
}

func (e *StatusEvent) Clone() StatusEvent {
	return StatusEvent{
		Type:  e.Type,
		Event: bytes.Clone(e.Event),
	}
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
	StatBackpackName,
	StatCargoName,
	StatFCMatsName,
	StatMarketName,
	StatModulesName,
	StatNavRouteName,
	StatOutfitName,
	StatLockerName,
	StatShipyardName,
	StatStatusName,
}
