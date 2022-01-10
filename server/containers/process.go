package containers

import "sync"

type Process struct {
	WordId       int
	Storyteller  int
	Teams        []uint
	Words        []string
	GuessedWords map[string]uint
	Mutex        *sync.RWMutex
}

func (p *Process) nextWord() (string, bool) {
	p.Mutex.RLock()
	defer p.Mutex.RUnlock()
	if len(p.Words) == len(p.GuessedWords) {
		return "", false
	}

	i := p.WordId
	for {
		_, ok := p.GuessedWords[p.Words[i]]
		if !ok {
			break
		} else {
			i = (i + 1) % len(p.Words)
		}
	}
	p.WordId = (i + 1) % len(p.Words)
	return p.Words[i], true
}

func (p *Process) guessWord(word string) {
	p.Mutex.Lock()
	defer p.Mutex.Unlock()
	p.GuessedWords[word] = p.Teams[p.Storyteller]
}
