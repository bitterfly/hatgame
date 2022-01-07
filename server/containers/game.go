package containers

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Game struct {
	Id         uint
	Players    MutexMap
	NumPlayers int
	Timer      int
	Host       uint
}

type Process struct {
	Teams        []uint
	GuessedWords map[string]uint
	Words        []string
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
	// TODO: remove 2
	if len(g.Players.Words[id]) == 2 {
		return nil, fmt.Errorf("words limit reached")
	}
	fmt.Printf("Adding %s to %d\n", word, id)
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

func (g *Game) CheckWordsFinished() bool {
	g.Players.WordsMutex.RLock()
	defer g.Players.WordsMutex.RUnlock()
	for _, w := range g.Players.Words {
		//TODO: remove 2 :D
		if len(w) != 2 {
			return false
		}
	}
	return true
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

func (g Game) StartProcess() *Process {
	teams := make([]uint, 0, len(g.Players.Ws))
	words := make([]string, 0)
	g.Players.WsMutex.RLock()
	for id, uwords := range g.Players.Words {
		teams = append(teams, id)
		for _, word := range uwords {
			words = append(words, word)
		}
	}
	g.Players.WsMutex.RUnlock()

	rand.Shuffle(
		len(teams),
		func(i, j int) { teams[i], teams[j] = teams[j], teams[i] },
	)

	rand.Shuffle(
		len(words),
		func(i, j int) { words[i], words[j] = words[j], words[i] },
	)

	return &Process{
		Teams: teams,
		Words: words,
	}

}

func (g *Game) Start(id uint) {
	// Make teams.
	process := g.StartProcess()

	fmt.Printf("Teams: %v\nWords: %v\n", process.Teams, process.Words)

	g.Players.WsMutex.RLock()
	for i, id := range process.Teams {
		resp := map[string]interface{}{
			"type":    "started",
			"partner": process.Teams[(i+int(float64(g.NumPlayers)/2))%g.NumPlayers],
		}

		respJson, err := json.Marshal(resp)
		if err != nil {
			fmt.Printf("Error when marshalling")
		}
		ws, _ := g.Players.Ws[id]
		err = ws.WriteMessage(websocket.TextMessage, respJson)
		if err != nil {
			fmt.Printf("Error when sending message")
		}
	}
	g.Players.WsMutex.RUnlock()
	// Send info about teams and game start

	// Pick storyteller
	storyteller := 0
	for {
		storyteller = storyteller % g.NumPlayers
		fmt.Printf("Storyteller: %d\n", process.Teams[storyteller])
		timer := time.NewTicker(1 * time.Second)
		done := make(chan struct{})
		go tick(g, done, timer)
		time.Sleep(5 * time.Second)
		timer.Stop()
		done <- struct{}{}

		storyteller += 1
	}
	// Select word at random
	// Tick timer

}

func tick(g *Game, done chan struct{}, timer *time.Ticker) {
	i := 0
	for {
		select {
		case <-done:
			return
		case <-timer.C:
			fmt.Printf("Tick: %d\n", i)
			i += 1
		}
	}
}
