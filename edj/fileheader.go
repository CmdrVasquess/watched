package edj

const FileheaderTag = "Fileheader"

type Fileheader struct {
	Event
	Part        int
	Language    string
	Odyssey     bool
	Gameversion string
	Build       string
}
