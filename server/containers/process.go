package containers

import "sync"

type Process struct {
	WordId       int
	Storyteller  int
	Teams        []uint
	GuessedWords map[string]uint
	Mutex        *sync.RWMutex
}

func (p *Process) guessWord(word string) {
	p.Mutex.Lock()
	defer p.Mutex.Unlock()
	p.GuessedWords[word] = p.Teams[p.Storyteller]
}
