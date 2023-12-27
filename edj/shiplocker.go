package edj

type ShipLocker struct {
	Event
	Items       []LockerItem
	Components  []LockerItem
	Consumables []LockerItem
	Data        []LockerItem
}

type LockerItem struct {
	Name          string
	NameLocalised string `json:"Name_Localised"`
	OwnerID       uint
	MissionID     uint
	Count         int
}
