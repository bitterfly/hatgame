package containers

import (
	"encoding/json"
	"fmt"
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
	Players    MutexMap
	Process    Process `json:"-"`
}

type Process struct {
	WordId       int
	Storyteller  int
	Teams        []uint
	Words        []string
	GuessedWords map[string]uint
	Mutex        *sync.RWMutex
}

type MutexMap struct {
	Ws         map[uint]*websocket.Conn
	Words      map[uint][]string
	WsMutex    *sync.RWMutex
	WordsMutex *sync.RWMutex
	Users      map[uint]schema.User
}

func (mm MutexMap) MarshalJSON() ([]byte, error) {
	Players := make([]schema.User, 0, len(mm.Users))
	for _, v := range mm.Users {
		Players = append(Players, v)
	}
	return json.Marshal(Players)
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

func NewGame(gameId uint, host schema.User, numPlayers, numWords, timer int) *Game {
	ws := make(map[uint]*websocket.Conn)
	ws[host.ID] = nil
	words := make(map[uint][]string)
	words[host.ID] = make([]string, 0)
	users := make(map[uint]schema.User)
	users[host.ID] = host

	return &Game{
		Id: gameId,
		Players: MutexMap{
			Ws:         ws,
			Words:      words,
			WsMutex:    &sync.RWMutex{},
			WordsMutex: &sync.RWMutex{},
			Users:      users},
		Process:    Process{},
		NumPlayers: numPlayers,
		NumWords:   numWords,
		Timer:      timer,
		Host:       host.ID,
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
	return g.CreateGameMessage()
}

func (g *Game) Get(id uint) (*websocket.Conn, bool) {
	g.Players.WsMutex.RLock()
	defer g.Players.WsMutex.RUnlock()
	ws, ok := g.Players.Ws[id]
	return ws, ok
}

func (g *Game) PutAll(max int, user schema.User, ws *websocket.Conn) ([]byte, error) {
	g.Players.WsMutex.Lock()
	defer g.Players.WsMutex.Unlock()
	if len(g.Players.Ws) == max {
		return nil, fmt.Errorf("too many players")
	}
	if _, ok := g.Players.Ws[user.ID]; ok {
		return nil, fmt.Errorf("player already in game")
	}
	fmt.Printf("Adding player with id: %d\n", user.ID)
	g.Players.Ws[user.ID] = ws
	g.Players.WordsMutex.Lock()
	defer g.Players.WordsMutex.Unlock()
	g.Players.Words[user.ID] = make([]string, 0)
	g.Players.Users[user.ID] = user
	return g.CreateGameMessage()
}

func (g *Game) AddWord(id uint, word string) ([]byte, error) {
	g.Players.WordsMutex.Lock()
	defer g.Players.WordsMutex.Unlock()
	if _, ok := g.Players.Words[id]; !ok {
		return nil, fmt.Errorf("no player with id %d", id)
	}
	if len(g.Players.Words[id]) == g.NumWords {
		return nil, fmt.Errorf("words limit reached")
	}
	fmt.Printf("Adding %s to %d\n", word, id)
	g.Players.Words[id] = append(g.Players.Words[id], word)
	return g.CreateWordMessage(word)
}

func (g Game) CreateWordMessage(word string) ([]byte, error) {
	msg := map[string]interface{}{
		"type": "word",
		"msg":  word,
	}

	return json.Marshal(msg)
}

func (g *Game) CheckWordsFinished() bool {
	g.Players.WordsMutex.RLock()
	defer g.Players.WordsMutex.RUnlock()
	for _, w := range g.Players.Words {
		if len(w) != g.NumWords {
			return false
		}
	}
	return true
}

func (g Game) CreateGameMessage() ([]byte, error) {
	msg := map[string]interface{}{
		"type": "game",
		"msg":  g,
	}

	return json.Marshal(msg)
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
		resp := map[string]interface{}{
			"type": "team",
			"msg":  g.Process.Teams[(i+int(float64(g.NumPlayers)/2))%g.NumPlayers],
		}

		respJson, err := json.Marshal(resp)
		if err != nil {
			return fmt.Errorf("error when marshalling team message: %w", err)
		}
		ws, _ := g.Players.Ws[id]
		err = ws.WriteMessage(websocket.TextMessage, respJson)
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

	resp := map[string]interface{}{
		"type": "end",
		"msg":  res,
	}
	respJson, err := json.Marshal(resp)
	if err != nil {
		return fmt.Errorf("error when marshalling end message: %w", err)
	}
	return game.NotifyAll(respJson)
}

func NotifyStoryteller(game *Game) error {
	resp := map[string]interface{}{
		"type": "start",
		"msg":  game.Process.Teams[game.Process.Storyteller],
	}
	respJson, err := json.Marshal(resp)
	if err != nil {
		return fmt.Errorf("error when marshalling start message: %w", err)
	}
	return game.NotifyAll(respJson)
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
	resp := map[string]interface{}{
		"type": "story",
		"msg":  story,
	}
	respJson, err := json.Marshal(resp)
	if err != nil {
		return err
	}

	game.Players.WsMutex.RLock()
	ws, _ := game.Players.Ws[game.Process.Teams[game.Process.Storyteller]]
	game.Players.WsMutex.RUnlock()

	return ws.WriteMessage(websocket.TextMessage, respJson)
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
			resp := map[string]interface{}{
				"type": "tick",
				"msg":  i,
			}
			respJson, err := json.Marshal(resp)
			if err != nil {
				fmt.Printf("Error when marshalling")
			}
			game.NotifyAll(respJson)
		}
	}
}
