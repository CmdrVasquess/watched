package edj

const LocationTag = "Location"

type Location struct {
	Event
	StarSystem                   string
	SystemAddress                uint64
	SystemAllegiance             string
	SystemEconomy                string
	SystemEconomyLocalised       string `json:"SystemEconomy_Localised"`
	SystemSecondEconomy          string
	SystemSecondEconomyLocalised string `json:"SystemSecondEconomy_Localised"`
	SystemGovernment             string
	SystemGovernmentLocalised    string `json:"SystemGovernment_Localised"`
	SystemSecurity               string
	SystemSecurityLocalised      string `json:"SystemSecurity_Localised"`
	Population                   uint
	Body                         string
	BodyID                       uint
	BodyType                     string
	Factions                     []Faction
	Latitude, Longitude          float32
	DistFromStarLS               float32
	StarPos                      [3]float32
	Docked                       bool
	Taxi                         bool
	Multicrew                    bool
}

type Faction struct {
	Name               string
	Influence          float32
	MyReputation       float32
	FactionState       string
	Government         string
	Allegiance         string
	Happiness          string
	HappinessLocalised string `json:"Happiness_Localised"`
	PendingStates      []FactionState
	RecoveringStates   []FactionState
	SystemFaction      struct{ Name string }
	Conflicts          []Conflict
}

type FactionState struct {
	State string
	Trend int
}

type Conflict struct {
	WarType            string
	Status             string
	Faction1, Faction2 struct {
		Name    string
		Stake   string
		WonDays int
	}
}
