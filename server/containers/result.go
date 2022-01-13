package containers

type Result struct {
	FirstID  uint
	SecondID uint
	Score    int
}

func (r Result) Contains(id uint) bool {
	return r.FirstID == id || r.SecondID == id
}

func Contains(rs []Result, id uint) bool {
	for _, r := range rs {
		if r.Contains(id) {
			return true
		}
	}
	return false
}
