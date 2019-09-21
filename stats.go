package watched

// Read: https://forums.frontier.co.uk/forums/elite-api-and-tools/
const (
	StatFlagFlagDocked uint32 = (1 << iota)
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
