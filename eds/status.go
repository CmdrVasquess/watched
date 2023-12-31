package eds

import "github.com/CmdrVasquess/watched/edj"

const StatusTag = "Status"

type Status struct {
	edj.Event
	Flags     StatusFlag
	Pips      [3]int8
	Firegroup int
	GuiFocus  int
	Fuel      struct{ FuelMain, FuelReservoir float64 }
	Cargo     float64
	// LegalState
	Latitude     float64
	Longitude    float64
	Altitude     int
	Heading      int
	BodyName     string
	PlanetRadius float64
}

func (s *Status) AnyFlag(fs StatusFlag) bool {
	return s.Flags&fs > 0
}

func (s *Status) AllFlags(fs StatusFlag) bool {
	return s.Flags&fs == fs
}

type StatusFlag = uint32

// Read: https://forums.frontier.co.uk/forums/elite-api-and-tools/
const (
	StatusDocked StatusFlag = (1 << iota)
	StatusLanded
	StatusGearDown
	StatusShieldsUp
	StatusSupercruise

	StatusFAOff
	StatusHPDeployed
	StatusInWing
	StatusLightsOn
	StatusCSDeployed

	StatusSilentRun
	StatusFuelScooping
	StatusSrvHandbrake
	StatusSrvTurret
	StatusSrvUnderShip

	StatusSrvDriveAssist
	StatusFsdMassLock
	StatusFsdCharging
	StatusCooldown
	StatusLowFuel

	StatusOverHeat
	StatusHasLatLon
	StatusIsInDanger
	StatusInterdicted
	StatusInMainShip

	StatusInFighter
	StatusInSrv
	StatusHudAnalysis
	StatusNightVis
	StatusAltAvgR

	StatusFSDJump
	StatusSrvHighBeam
)

type GuiFocus = int

const (
	StatusNoFocus GuiFocus = iota
	StatusInternalPanel
	StatusExternalPanel
	StatusCommsPanel
	StatusRolePanel
	StatusStationServices
	StatusGalaxyMap
	StatusSystemMap
	StatusOrrery
	StatusFSSMode
	StatusSAAMode
	StatusCodex
)
