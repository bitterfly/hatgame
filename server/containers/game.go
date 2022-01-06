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
	Ws         map[uint]*websocket.Conn
	Words      map[uint][]string
	WsMutex    *sync.RWMutex
	WordsMutex *sync.RWMutex
}

func (mm MutexMap) MarshalJSON() ([]byte, error) {
	Players := make([]uint, 0, len(mm.Ws))
	for k := range mm.Ws {
		Players = append(Players, k)
	}
	return json.Marshal(Players)
}

func NewGame(id, host uint, numPlayers, timer int) *Game {
	ws := make(map[uint]*websocket.Conn)
	ws[host] = nil
	words := make(map[uint][]string)
	words[host] = make([]string, 0)

	return &Game{
		Id: id,
		Players: MutexMap{
			Ws:         ws,
			Words:      words,
			WsMutex:    &sync.RWMutex{},
			WordsMutex: &sync.RWMutex{}},
		NumPlayers: numPlayers,
		Timer:      timer,
		Host:       host,
	}
}

func (g *Game) Put(max int, id uint) error {
	g.Players.WsMutex.Lock()
	defer g.Players.WsMutex.Unlock()
	if len(g.Players.Ws) == max {
		return fmt.Errorf("too many players")
	}
	if _, ok := g.Players.Ws[id]; ok {
		return fmt.Errorf("player already in game")
	}
	fmt.Printf("Adding player with id: %d\n", id)
	g.Players.Ws[id] = nil
	g.Players.WordsMutex.Lock()
	defer g.Players.WordsMutex.Unlock()
	g.Players.Words[id] = make([]string, 0)
	return nil
}

func (g *Game) PutWs(id uint, ws *websocket.Conn) ([]byte, error) {
	if _, ok := g.Players.Ws[id]; !ok {
		return nil, fmt.Errorf("no such player in game")
	}
	g.Players.WsMutex.Lock()
	defer g.Players.WsMutex.Unlock()
	g.Players.Ws[id] = ws
	if len(g.Players.Ws) == g.NumPlayers {
		return g.CreateGameMessage("done")
	}
	return g.CreateGameMessage("ok")
}

func (g *Game) Get(id uint) (*websocket.Conn, bool) {
	g.Players.WsMutex.RLock()
	defer g.Players.WsMutex.RUnlock()
	ws, ok := g.Players.Ws[id]
	return ws, ok
}

func (g *Game) PutAll(max int, id uint, ws *websocket.Conn) ([]byte, error) {
	g.Players.WsMutex.Lock()
	defer g.Players.WsMutex.Unlock()
	if len(g.Players.Ws) == max {
		return nil, fmt.Errorf("too many players")
	}
	if _, ok := g.Players.Ws[id]; ok {
		return nil, fmt.Errorf("player already in game")
	}
	fmt.Printf("Adding player with id: %d\n", id)
	g.Players.Ws[id] = ws
	g.Players.WordsMutex.Lock()
	defer g.Players.WordsMutex.Unlock()
	g.Players.Words[id] = make([]string, 0)
	if len(g.Players.Ws) == g.NumPlayers {
		return g.CreateGameMessage("done")
	}
	return g.CreateGameMessage("ok")
}

func (g *Game) AddWord(id uint, word string) ([]byte, error) {
	g.Players.WordsMutex.Lock()
	defer g.Players.WordsMutex.Unlock()
	if _, ok := g.Players.Words[id]; !ok {
		return nil, fmt.Errorf("no player with id %d", id)
	}
	if len(g.Players.Words[id]) == 2 {
		return g.CreateWordMessage(word, "done")
	}
	g.Players.Words[id] = append(g.Players.Words[id], word)
	return g.CreateWordMessage(word, "ok")
}

func (g Game) CreateWordMessage(word string, status string) ([]byte, error) {
	msg := map[string]interface{}{
		"type":   "word",
		"status": status,
		"msg":    word,
	}

	return json.Marshal(msg)
}

func (g Game) CreateGameMessage(status string) ([]byte, error) {
	msg := map[string]interface{}{
		"type":   "game",
		"status": status,
		"msg":    g,
	}

	return json.Marshal(msg)
}

func (g Game) WriteAll(msg []byte) error {

	g.Players.WsMutex.RLock()
	defer g.Players.WsMutex.RUnlock()
	for _, ws := range g.Players.Ws {
		err := ws.WriteMessage(websocket.TextMessage, msg)
		if err != nil {
			return err
		}
	}
	return nil
}
