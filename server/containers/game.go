package containers

import (
	"encoding/json"
	"fmt"
	"sync"
)

type Game struct {
	Id         uint
	Players    MutexMap
	NumPlayers int
	Timer      int
	Host       uint
}

type MutexMap struct {
	Data  map[uint]struct{}
	Mutex *sync.RWMutex
}

func (mm MutexMap) MarshalJSON() ([]byte, error) {
	Players := make([]uint, len(mm.Data))
	for k := range mm.Data {
		Players = append(Players, k)
	}
	return json.Marshal(Players)
}

func NewGame(id, host uint, players, timer int) Game {
	playerIds := make(map[uint]struct{})
	playerIds[host] = struct{}{}

	return Game{
		Id: id,
		Players: MutexMap{
			Data:  playerIds,
			Mutex: &sync.RWMutex{}},
		NumPlayers: players,
		Timer:      timer,
		Host:       host,
	}
}

func (g *Game) Put(max int, id uint) error {
	g.Players.Mutex.Lock()
	defer g.Players.Mutex.Unlock()
	if len(g.Players.Data) == max {
		return fmt.Errorf("too many players")
	}
	if _, ok := g.Players.Data[id]; ok {
		return fmt.Errorf("player already in game")
	}
	fmt.Printf("Adding player with id: %d\n", id)
	g.Players.Data[id] = struct{}{}
	return nil
}

func (g *Game) Get(id uint) bool {
	g.Players.Mutex.RLock()
	defer g.Players.Mutex.RUnlock()
	_, ok := g.Players.Data[id]
	return ok
}
