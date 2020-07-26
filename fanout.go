package watched

type Branch struct {
	EDEvents
	j chan JounalEvent
	s chan StatusEvent
}

type FanOut struct {
	Input *EDEvents
	out   []Branch
}

func (f *FanOut) Branch(qlen int) *Branch {
	i := len(f.out)
	f.out = append(f.out, Branch{
		j: make(chan JounalEvent, qlen),
		s: make(chan StatusEvent, qlen),
	})
	res := &f.out[i]
	res.Journal = res.j
	res.Status = res.s
	return res
}

func (f *FanOut) Run() {
	open := 0
	if f.Input.Journal != nil {
		open++
	}
	if f.Input.Status != nil {
		open++
	}
	for open > 0 {
		select {
		case je, ok := <-f.Input.Journal:
			if ok {
				for _, out := range f.out {
					out.j <- je
				}
			} else {
				open--
			}
		case se, ok := <-f.Input.Status:
			if ok {
				for _, out := range f.out {
					out.s <- se
				}
			} else {
				open--
			}
		}
	}
	for i := range f.out {
		close(f.out[i].j)
		close(f.out[i].s)
	}
}
