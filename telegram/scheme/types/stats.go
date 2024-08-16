package types

type Stats struct {
	Additions int
	Changes   int
}

func (s *Stats) Total() int {
	return s.Additions + s.Changes
}
