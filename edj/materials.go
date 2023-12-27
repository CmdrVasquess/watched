package edj

const MaterialsTag = "Materials"

type Materials struct {
	Event
	Raw          []RawMaterial
	Manufactured []Material
	Encoded      []Material
}

type RawMaterial struct {
	Name  string
	Count int
}

type Material struct {
	RawMaterial
	NameLocalised string `json:"Name_Localised"`
}
