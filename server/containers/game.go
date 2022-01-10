package containers

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"sort"
	"sync"
	"time"

	"github.com/bitterfly/go-chaos/hatgame/schema"
	"github.com/gorilla/websocket"
)

type Game struct {
	Id         uint
	Host       uint
	NumPlayers int
	Timer      int
	NumWords   int
	Players    Players
	Process    Process `json:"-"`
}

type Players struct {
	Ws          map[uint]*websocket.Conn
	WordsByUser map[uint]map[string]struct{}
	Words       map[string]struct{}
	WsMutex     *sync.RWMutex
	WordsMutex  *sync.RWMutex
	Users       map[uint]schema.User
}

func (p Players) MarshalJSON() ([]byte, error) {
	Players := make([]schema.User, 0, len(p.Users))
	for _, v := range p.Users {
		Players = append(Players, v)
	}
	return json.Marshal(Players)
}

func NewGame(gameId uint, host schema.User, numPlayers, numWords, timer int) *Game {
	ws := make(map[uint]*websocket.Conn)
	ws[host.ID] = nil
	wordsByUser := make(map[uint]map[string]struct{})
	wordsByUser[host.ID] = make(map[string]struct{})
	users := make(map[uint]schema.User)
	users[host.ID] = host
	words := make(map[string]struct{})

	return &Game{
		Id: gameId,
		Players: Players{
			Ws:          ws,
			WordsByUser: wordsByUser,
			Words:       words,
			WsMutex:     &sync.RWMutex{},
			WordsMutex:  &sync.RWMutex{},
			Users:       users},
		Process:    Process{},
		NumPlayers: numPlayers,
		NumWords:   numWords,
		Timer:      timer,
		Host:       host.ID,
	}
}

func (g *Game) Get(id uint) (*websocket.Conn, bool) {
	g.Players.WsMutex.RLock()
	defer g.Players.WsMutex.RUnlock()
	ws, ok := g.Players.Ws[id]
	return ws, ok
}

func (g *Game) PutWs(id uint, ws *websocket.Conn) error {
	if _, ok := g.Players.Ws[id]; !ok {
		return fmt.Errorf("no such player in game")
	}
	g.Players.WsMutex.Lock()
	defer g.Players.WsMutex.Unlock()
	g.Players.Ws[id] = ws
	return nil
}

func (g *Game) Put(max int, user schema.User, ws *websocket.Conn) error {
	g.Players.WsMutex.Lock()
	defer g.Players.WsMutex.Unlock()
	if len(g.Players.Ws) == max {
		return fmt.Errorf("too many players")
	}
	if _, ok := g.Players.Ws[user.ID]; ok {
		return fmt.Errorf("player already in game")
	}
	log.Printf("Adding player with id: %d\n", user.ID)
	g.Players.Ws[user.ID] = ws
	g.Players.WordsMutex.Lock()
	defer g.Players.WordsMutex.Unlock()
	g.Players.WordsByUser[user.ID] = make(map[string]struct{})
	g.Players.Users[user.ID] = user
	return nil
}

func (g *Game) AddWord(id uint, word string) ([]byte, error) {
	g.Players.WordsMutex.Lock()
	defer g.Players.WordsMutex.Unlock()
	if _, ok := g.Players.WordsByUser[id]; !ok {
		return nil, fmt.Errorf("no player with id %d", id)
	}
	if len(g.Players.WordsByUser[id]) == g.NumWords {
		return nil, fmt.Errorf("words limit reached")
	}
	if _, ok := g.Players.Words[word]; ok {
		return CreateMessage("error", "Already used this word")
	}
	fmt.Printf("Adding %s to %d\n", word, id)
	g.Players.WordsByUser[id][word] = struct{}{}
	g.Players.Words[word] = struct{}{}
	return CreateMessage("word", word)
}

func (g *Game) CheckWordsFinished() bool {
	g.Players.WordsMutex.RLock()
	defer g.Players.WordsMutex.RUnlock()
	return len(g.Players.Words) == g.NumPlayers*g.NumWords
}

func (g Game) NotifyAll(msg []byte) error {
	fmt.Printf("Writing: %s\n", msg)

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

func (g *Game) StartProcess() {
	teams := make([]uint, 0, len(g.Players.Ws))
	words := make([]string, 0, len(g.Players.Words))
	g.Players.WsMutex.RLock()
	for id := range g.Players.WordsByUser {
		teams = append(teams, id)

	}
	for word := range g.Players.Words {
		words = append(words, word)
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

	g.Process = Process{
		Teams:        teams,
		Words:        words,
		GuessedWords: make(map[string]uint),
		Mutex:        &sync.RWMutex{},
		Storyteller:  0,
		WordId:       0,
	}

}

func NotifyGameStarted(g *Game) error {
	g.Players.WsMutex.RLock()
	for i, id := range g.Process.Teams {
		resp, err := CreateMessage("team",
			g.Process.Teams[(i+int(float64(g.NumPlayers)/2))%g.NumPlayers])
		if err != nil {
			return fmt.Errorf("error when marshalling team message: %w", err)
		}
		ws, _ := g.Players.Ws[id]
		err = ws.WriteMessage(websocket.TextMessage, resp)
		if err != nil {
			return fmt.Errorf("error when sending team message: %w", err)
		}
	}
	g.Players.WsMutex.RUnlock()
	return nil
}

type Team struct {
	First  uint
	Second uint
	Score  int
}

func makeTeam(a, b uint) Team {
	return Team{First: a, Second: b}
}

func NotifyGameEnded(game *Game) error {
	fmt.Printf("Notify game end\n")

	game.Process.Mutex.RLock()
	defer game.Process.Mutex.RUnlock()

	rev := make(map[uint]int)
	for _, id := range game.Process.GuessedWords {
		rev[id] += 1
	}
	teams := int(game.NumPlayers / 2.0)
	res := make([]Team, 0, teams)
	for i := 0; i < teams; i++ {
		team := makeTeam(
			game.Process.Teams[i],
			game.Process.Teams[(i+teams)%game.NumPlayers],
		)
		team.Score =
			rev[team.First] + rev[team.Second]
		res = append(res, team)
	}

	sort.SliceStable(res, func(i, j int) bool {
		return res[i].Score > res[j].Score
	})

	resp, err := CreateMessage("end", res)
	if err != nil {
		return fmt.Errorf("error when marshalling end message: %w", err)
	}
	return game.NotifyAll(resp)
}

func NotifyStoryteller(game *Game) error {
	resp, err := CreateMessage("start",
		game.Process.Teams[game.Process.Storyteller])
	if err != nil {
		return fmt.Errorf("error when marshalling start message: %w", err)
	}
	return game.NotifyAll(resp)
}

func Start(id uint, game *Game) error {
	game.StartProcess()
	fmt.Printf("Teams: %v\nWords: %v\n", game.Process.Teams, game.Process.Words)
	err := NotifyGameStarted(game)
	if err != nil {
		return err
	}
	return NotifyStoryteller(game)
}

func NotifyWord(game *Game, story string) error {
	resp, err := CreateMessage("story", story)
	if err != nil {
		return err
	}

	game.Players.WsMutex.RLock()
	ws, _ := game.Players.Ws[game.Process.Teams[game.Process.Storyteller]]
	game.Players.WsMutex.RUnlock()

	return ws.WriteMessage(websocket.TextMessage, resp)
}

func PickWord(game *Game) (bool, error) {
	story, found := game.Process.nextWord()
	fmt.Printf("Story chosen: %s\n", story)

	if !found {
		return false, NotifyGameEnded(game)
	}

	err := NotifyWord(game, story)
	if err != nil {
		return true, err
	}
	return true, nil
}

func MakeTurn(id uint, game *Game, timerGameEnd chan struct{}) error {
	ok, err := PickWord(game)
	if !ok || err != nil {
		return err
	}

	timer := time.NewTicker(1 * time.Second)
	timerDone := make(chan struct{})
	go tick(game, timerDone, timerGameEnd, timer)
	time.Sleep(time.Duration(game.Timer) * time.Second)
	timer.Stop()
	timerDone <- struct{}{}

	game.Process.Storyteller = (game.Process.Storyteller + 1) % game.NumPlayers
	return NotifyStoryteller(game)
}

func tick(game *Game, timerDone chan struct{}, timerGameEnd chan struct{}, timer *time.Ticker) {
	i := game.Timer
	for {
		select {
		case <-timerDone:
			fmt.Printf("Timer done\n")
			return
		case <-timerGameEnd:
			fmt.Printf("Game end done\n")
			return
		case _, ok := <-timer.C:
			if !ok {
				return
			}
			i -= 1
			fmt.Printf("Tick: %d\n", i)

			resp, err := CreateMessage("tick", i)
			if err != nil {
				fmt.Printf("Error when marshalling")
			}
			game.NotifyAll(resp)
		}
	}
}
