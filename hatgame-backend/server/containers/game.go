package containers

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"sort"
	"sync"
	"time"

	"github.com/bitterfly/go-chaos/hatgame/schema"
	"github.com/bitterfly/go-chaos/hatgame/utils"
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

func (g *Game) nextWord() (string, bool) {
	g.Process.Mutex.RLock()
	defer g.Process.Mutex.RUnlock()
	g.Players.WordsMutex.RLock()
	defer g.Players.WordsMutex.RUnlock()

	if len(g.Players.Words) == len(g.Process.GuessedWords) {
		return "", false
	}

	unguessed := make([]string, 0, len(g.Players.Words)-len(g.Process.GuessedWords))
	for word := range g.Players.Words {
		if _, ok := g.Process.GuessedWords[word]; !ok {
			unguessed = append(unguessed, word)
		}
	}

	return unguessed[rand.Intn(len(unguessed))], true
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
		Process: Process{
			Teams:        make([]uint, 0, numPlayers),
			GuessedWords: make(map[string]uint),
			Mutex:        &sync.RWMutex{},
			GameEnd:      make(chan struct{}),
			Storyteller:  0,
			WordId:       0,
		},
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

func (g *Game) MakeTeams() {
	g.Players.WordsMutex.RLock()
	for id := range g.Players.WordsByUser {
		g.Process.Teams = append(g.Process.Teams, id)

	}
	g.Players.WordsMutex.RUnlock()

	rand.Shuffle(
		len(g.Process.Teams),
		func(i, j int) {
			g.Process.Teams[i], g.Process.Teams[j] = g.Process.Teams[j], g.Process.Teams[i]
		},
	)
}

func NotifyGameStarted(g *Game) error {
	g.Players.WsMutex.RLock()
	for i, id := range g.Process.Teams {
		resp, err := CreateMessage("team",
			g.Process.Teams[(i+int(float64(g.NumPlayers)/2))%g.NumPlayers])
		if err != nil {
			return fmt.Errorf("error when marshalling team message: %w", err)
		}
		ws := g.Players.Ws[id]
		err = ws.WriteMessage(websocket.TextMessage, resp)
		if err != nil {
			return fmt.Errorf("error when sending team message: %w", err)
		}
	}
	g.Players.WsMutex.RUnlock()
	return nil
}

func NotifyGameEnded(game *Game) error {
	game.Process.Mutex.RLock()
	defer game.Process.Mutex.RUnlock()

	rev := make(map[uint]int)
	for _, id := range game.Process.GuessedWords {
		rev[id] += 1
	}
	teams := int(game.NumPlayers / 2.0)
	game.Process.Result = make([]Result, 0, teams)

	for i := 0; i < teams; i++ {
		first, second := utils.Order(
			game.Process.Teams[i],
			game.Process.Teams[(i+teams)%game.NumPlayers])

		res := Result{FirstID: first, SecondID: second}
		res.Score =
			rev[res.FirstID] + rev[res.SecondID]
		game.Process.Result = append(game.Process.Result, res)

	}

	sort.SliceStable(game.Process.Result, func(i, j int) bool {
		return game.Process.Result[i].Score > game.Process.Result[j].Score
	})

	resp, err := CreateMessage("end", game.Process.Result)
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

func (game *Game) Start(id uint) error {
	game.MakeTeams()
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
	ws := game.Players.Ws[game.Process.Teams[game.Process.Storyteller]]
	game.Players.WsMutex.RUnlock()

	return ws.WriteMessage(websocket.TextMessage, resp)
}

func MakeTurn(id uint, game *Game) error {
	story, found := game.nextWord()

	if !found {
		return NotifyGameEnded(game)
	}

	err := NotifyWord(game, story)
	if err != nil {
		return err
	}

	timer := time.NewTicker(1 * time.Second)
	timerDone := make(chan struct{})
	go tick(game, timerDone, timer)
	go func(timerDone chan struct{}) {
		time.Sleep(time.Duration(game.Timer) * time.Second)
		timer.Stop()
		close(timerDone)
	}(timerDone)

	select {
	case _, ok := <-game.Process.GameEnd:
		if !ok {
			return nil
		}
	case _, ok := <-timerDone:
		if !ok {
			game.Process.Storyteller = (game.Process.Storyteller + 1) % game.NumPlayers
			return NotifyStoryteller(game)
		}
	}
	return nil
}

func tick(game *Game, timerDone chan struct{}, timer *time.Ticker) {
	i := game.Timer
	defer fmt.Printf("Closing tick\n")

	for {
		select {
		case <-timerDone:
			return
		case _, ok := <-game.Process.GameEnd:
			if !ok {
				return
			}
		case _, ok := <-timer.C:
			if !ok {
				return
			}
			i -= 1
			// fmt.Printf("Tick: %d\n", i)

			resp, err := CreateMessage("tick", i)
			if err != nil {
				fmt.Printf("Error when marshalling")
			}
			game.NotifyAll(resp)
		}
	}
}
