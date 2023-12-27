package edj

type LoadGame struct {
	Event
	FID           string
	Commander     string
	Horizons      bool
	Odyssey       bool
	Ship          string
	ShipLocalised string `json:"Ship_Localised"`
	ShipID        uint
	ShipName      string
	ShipIdent     string
	FuelLevel     float32
	FuelCapacity  float32
	GameMode      string
	Credits       uint64
	Loan          uint64
	Language      string
	Gameversion   string
	Build         string
}
