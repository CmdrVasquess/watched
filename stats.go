package watched

// Read: https://forums.frontier.co.uk/forums/elite-api-and-tools/
const (
	StatFlagDocked uint32 = (1 << iota)
	StatFlagLanded
	StatFlagGearDown
	StatFlagShieldsUp
	StatFlagSupercruise

	StatFlagFAOff
	StatFlagHPDeployed
	StatFlagInWing
	StatFlagLightsOn
	StatFlagCSDeployed

	StatFlagSilentRun
	StatFlagFuelScooping
	StatFlagSrvHandbrake
	StatFlagSrvTurret
	StatFlagSrvUnderShip

	StatFlagSrvDriveAssist
	StatFlagFsdMassLock
	StatFlagFsdCharging
	StatFlagCooldown
	StatFlagLowFuel

	StatFlagOverHeat
	StatFlagHasLatLon
	StatFlagIsInDanger
	StatFlagInterdicted
	StatFlagInMainShip

	StatFlagInFighter
	StatFlagInSrv
	StatFlagHudAnalysis
	StatFlagNightVis
	StatFlagAltAvgR
)

func FlagsAny(set, test uint32) bool { return set&test != 0 }

func FlagsAll(set, test uint32) bool { return set%test == test }
