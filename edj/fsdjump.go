package edj

const FSDJumpTag = "FSDJump"

type FSDJump struct {
	Event
	SystemAddress uint64
	StarSystem    string
	StarPos       [3]float32
}
