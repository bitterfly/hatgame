package containers

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/gorilla/websocket"
)

type Game struct {
	Id         uint
	Players    MutexMap
	NumPlayers int
	Timer      int
	Host       uint
}

type MutexMap struct {
	Data  map[uint]*websocket.Conn
	Mutex *sync.RWMutex
}

func (mm MutexMap) MarshalJSON() ([]byte, error) {
	Players := make([]uint, 0, len(mm.Data))
	for k := range mm.Data {
		Players = append(Players, k)
	}
	return json.Marshal(Players)
}

func NewGame(id, host uint, numPlayers, timer int) *Game {
	playerIds := make(map[uint]*websocket.Conn)
	playerIds[host] = nil

	return &Game{
		Id: id,
		Players: MutexMap{
			Data:  playerIds,
			Mutex: &sync.RWMutex{}},
		NumPlayers: numPlayers,
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
	g.Players.Data[id] = nil
	return nil
}

func (g *Game) PutWs(id uint, ws *websocket.Conn) error {
	if _, ok := g.Players.Data[id]; !ok {
		return fmt.Errorf("no such player in game")
	}
	g.Players.Data[id] = ws
	return nil
}

func (g *Game) Get(id uint) (*websocket.Conn, bool) {
	g.Players.Mutex.RLock()
	defer g.Players.Mutex.RUnlock()
	ws, ok := g.Players.Data[id]
	return ws, ok
}

func (g *Game) PutAll(max int, id uint, ws *websocket.Conn) error {
	g.Players.Mutex.Lock()
	defer g.Players.Mutex.Unlock()
	if len(g.Players.Data) == max {
		return fmt.Errorf("too many players")
	}
	if _, ok := g.Players.Data[id]; ok {
		return fmt.Errorf("player already in game")
	}
	fmt.Printf("Adding player with id: %d\n", id)
	g.Players.Data[id] = ws
	return nil
}

func (g Game) CreateMessage() ([]byte, error) {
	msg := map[string]interface{}{
		"type": "game",
		"msg":  g,
	}

	return json.Marshal(msg)
}

func (g Game) WriteAll() error {
	msg, err := g.CreateMessage()
	if err != nil {
		return err
	}

	g.Players.Mutex.RLock()
	defer g.Players.Mutex.RUnlock()
	for _, ws := range g.Players.Data {
		err = ws.WriteMessage(websocket.TextMessage, msg)
		if err != nil {
			return err
		}
	}
	return nil
}
