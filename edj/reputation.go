package edj

const ReputationTag = "Reputation"

type Reputation struct {
	Event
	Empire, Federation, Independent, Alliance float64
}
