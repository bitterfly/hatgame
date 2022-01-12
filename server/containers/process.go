package containers

import "sync"

type Process struct {
	WordId       int
	Storyteller  int
	Teams        []uint
	WinningTeam  Team
	GuessedWords map[string]uint
	Mutex        *sync.RWMutex
	GameEnd      chan struct{}
}

func (p *Process) guessWord(word string) {
	p.Mutex.Lock()
	defer p.Mutex.Unlock()
	p.GuessedWords[word] = p.Teams[p.Storyteller]
}
