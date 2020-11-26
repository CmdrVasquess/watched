package watched

type branch struct {
	EventSrc
	j chan JounalEvent
	s chan StatusEvent
}

type RecvToSrc struct {
	out []branch
}

type BranchConfig struct {
	JournalQLen int
	StatusQLen  int
}

func (rs *RecvToSrc) Branch(cfg BranchConfig) EventSrc {
	i := len(rs.out)
	rs.out = append(rs.out, branch{
		j: make(chan JounalEvent, cfg.JournalQLen),
		s: make(chan StatusEvent, cfg.StatusQLen),
	})
	res := &rs.out[i]
	res.Journal = res.j
	res.Status = res.s
	return res.EventSrc
}

func (rs *RecvToSrc) Journal(e JounalEvent) error {
	for _, b := range rs.out {
		b.j <- e
	}
	return nil
}

func (rs *RecvToSrc) Status(e StatusEvent) error {
	for _, b := range rs.out {
		b.s <- e
	}
	return nil
}

func (rs *RecvToSrc) Close() error {
	for _, b := range rs.out {
		close(b.j)
		close(b.s)
	}
	rs.out = nil
	return nil
}
