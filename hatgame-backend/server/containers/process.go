package containers

import (
	"sort"
	"sync"

	"github.com/bitterfly/go-chaos/hatgame/utils"
)

type Process struct {
	WordId       int
	Storyteller  int
	Teams        []uint
	Result       []Result
	GuessedWords map[string]uint
	Mutex        *sync.RWMutex
	GameEnd      chan struct{}
}

func (p *Process) guessWord(word string) {
	p.Mutex.Lock()
	defer p.Mutex.Unlock()
	p.GuessedWords[word] = p.Teams[p.Storyteller]
}

func (p *Process) GetResults() {
	teams := int(len(p.Teams) / 2.0)
	rev := make(map[uint]int)
	for _, id := range p.GuessedWords {
		rev[id] += 1
	}
	p.Result = make([]Result, 0, teams)

	for i := 0; i < teams; i++ {
		first, second := utils.Order(
			p.Teams[i],
			p.Teams[(i+teams)%len(p.Teams)])

		res := Result{FirstID: first, SecondID: second}
		res.Score =
			rev[res.FirstID] + rev[res.SecondID]
		p.Result = append(p.Result, res)

	}

	sort.SliceStable(p.Result, func(i, j int) bool {
		return p.Result[i].Score > p.Result[j].Score
	})
}
