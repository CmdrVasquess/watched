package eds

import "github.com/CmdrVasquess/watched/edj"

const NavRouteTag = "NavRoute"

type NavRoute struct {
	edj.Event
	Route []NavSystem
}

type NavSystem struct {
	StarSystem    string
	SystemAddress uint64
	StarPos       [3]float32
	StarClass     string
}
