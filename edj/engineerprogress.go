package edj

type EngineerProgress struct {
	Event
	Engineers []Engineer
}

type Engineer struct {
	Engineer     string
	EngineerID   uint
	Progress     string
	RankProgress int
	Rank         int
}
