package containers

import (
	"fmt"
	"sync"
)

type Game struct {
	Players    map[uint]struct{}
	NumPlayers int
	Timer      int
	Host       uint
	Mutex      *sync.RWMutex `json:"-"`
}

func NewGame(host uint, players, timer int) Game {
	playerIds := make(map[uint]struct{})
	playerIds[host] = struct{}{}

	return Game{
		Players:    playerIds,
		NumPlayers: players,
		Timer:      timer,
		Host:       host,
		Mutex:      &sync.RWMutex{},
	}
}

func (g *Game) Put(max int, id uint) error {
	g.Mutex.Lock()
	defer g.Mutex.Unlock()
	if len(g.Players) == max {
		return fmt.Errorf("too many players")
	}
	if _, ok := g.Players[id]; ok {
		return fmt.Errorf("player already in game")
	}
	fmt.Printf("Adding player with id: %d\n", id)
	g.Players[id] = struct{}{}
	return nil
}

func (g *Game) Get(id uint) bool {
	g.Mutex.RLock()
	defer g.Mutex.RUnlock()
	_, ok := g.Players[id]
	return ok
}
