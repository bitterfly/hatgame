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
)

type MessageType string

const (
	Tick     MessageType = "tick"
	Ready    MessageType = "ready"
	AddWord  MessageType = "word"
	Guess    MessageType = "guess"
	GameInfo MessageType = "game"
	Team     MessageType = "team"
	End      MessageType = "end"
	Start    MessageType = "start"
	Story    MessageType = "story"
	Error    MessageType = "error"
)

type Message struct {
	Type MessageType
	Msg  interface{}
}

type Event struct {
	Type      MessageType
	Msg       interface{}
	Receivers []uint
	GameID    uint
}

type Game struct {
	ID         uint
	Host       uint
	NumPlayers int
	Timer      int
	NumWords   int
	Players    Players
	Process    Process    `json:"-"`
	Events     chan Event `json:"-"`
	Errors     chan error `json:"-"`

	PlayersIDs []uint `json:"-"`
}

type Players struct {
	WordsByUser map[uint]map[string]struct{}
	Words       map[string]struct{}
	WordsMutex  *sync.RWMutex
	Users       map[uint]schema.User
}

type Process struct {
	WordId       int
	Storyteller  int
	Teams        []uint
	Result       []Result
	GuessedWords map[string]uint
	Mutex        *sync.RWMutex
	GameEnd      chan struct{}
}

func (g *Game) GuessWord(word string) {
	g.Process.Mutex.Lock()
	defer g.Process.Mutex.Unlock()
	g.Process.GuessedWords[word] = g.Process.Teams[g.Process.Storyteller]
}

func (g *Game) GetResults() {
	teams := int(len(g.Process.Teams) / 2.0)
	rev := make(map[uint]int)
	for _, id := range g.Process.GuessedWords {
		rev[id] += 1
	}
	g.Process.Result = make([]Result, 0, teams)

	for i := 0; i < teams; i++ {
		first, second := utils.Order(
			g.Process.Teams[i],
			g.Process.Teams[(i+teams)%len(g.Process.Teams)])

		res := Result{FirstID: first, SecondID: second}
		res.Score =
			rev[res.FirstID] + rev[res.SecondID]
		g.Process.Result = append(g.Process.Result, res)

	}

	sort.SliceStable(g.Process.Result, func(i, j int) bool {
		return g.Process.Result[i].Score > g.Process.Result[j].Score
	})
}

func (g *Game) GetNextWord() {
	word, found := g.nextWord()
	if !found {
		NotifyGameEnded(g)
		close(g.Process.GameEnd)
		close(g.Events)
		close(g.Errors)
		return
	}
	NotifyWord(g, word)
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
	wordsByUser := make(map[uint]map[string]struct{})
	wordsByUser[host.ID] = make(map[string]struct{})
	users := make(map[uint]schema.User)
	users[host.ID] = host
	words := make(map[string]struct{})
	rand.Seed(time.Now().UnixNano())

	return &Game{
		ID: gameId,
		Players: Players{
			WordsByUser: wordsByUser,
			Words:       words,
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
		Events:     make(chan Event),
		Errors:     make(chan error),
		PlayersIDs: []uint{host.ID},
	}
}

func (g *Game) AddPlayer(max int, user schema.User) error {
	if len(g.Players.Users) == max {
		return fmt.Errorf("too many players")
	}
	if _, ok := g.Players.Users[user.ID]; ok {
		return fmt.Errorf("player already in game")
	}
	g.Players.WordsMutex.Lock()
	defer g.Players.WordsMutex.Unlock()
	g.Players.WordsByUser[user.ID] = make(map[string]struct{})
	g.Players.Users[user.ID] = user
	g.PlayersIDs = append(g.PlayersIDs, user.ID)
	return nil
}

func (g *Game) addWord(id uint, word string) (bool, error) {
	g.Players.WordsMutex.Lock()
	defer g.Players.WordsMutex.Unlock()
	if _, ok := g.Players.WordsByUser[id]; !ok {
		return false, fmt.Errorf("no player with id %d", id)
	}
	if len(g.Players.WordsByUser[id]) == g.NumWords {
		return false, fmt.Errorf("words limit reached")
	}
	if _, ok := g.Players.Words[word]; ok {
		return false, nil
	}
	g.Players.WordsByUser[id][word] = struct{}{}
	g.Players.Words[word] = struct{}{}
	return true, nil
}

func (g *Game) AddWord(id uint, word string) {
	ok, err := g.addWord(id, word)
	g.Errors <- err

	if !ok {
		g.Events <- Event{
			GameID:    g.ID,
			Type:      Error,
			Msg:       "Already used this word",
			Receivers: []uint{id},
		}
		return
	}

	g.Events <- Event{
		GameID:    g.ID,
		Type:      AddWord,
		Msg:       word,
		Receivers: []uint{id},
	}

	if g.CheckWordsFinished() {
		g.MakeTeams()
		NotifyGameStarted(g)
		NotifyStoryteller(g)
	}
}

func (g *Game) CheckWordsFinished() bool {
	g.Players.WordsMutex.RLock()
	defer g.Players.WordsMutex.RUnlock()
	return len(g.Players.Words) == g.NumPlayers*g.NumWords
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

func NotifyGameStarted(g *Game) {
	for i, id := range g.Process.Teams {
		g.Events <- Event{
			GameID:    g.ID,
			Type:      Team,
			Msg:       g.Process.Teams[(i+int(float64(g.NumPlayers)/2))%g.NumPlayers],
			Receivers: []uint{id},
		}
	}
}

func NotifyGameInfo(game *Game) {
	game.Events <- Event{
		GameID:    game.ID,
		Type:      GameInfo,
		Msg:       game,
		Receivers: game.PlayersIDs,
	}
}

func NotifyGameEnded(game *Game) {
	game.GetResults()
	game.Events <- Event{
		GameID:    game.ID,
		Receivers: game.PlayersIDs,
		Type:      End,
		Msg:       game.Process.Result,
	}
}

func NotifyStoryteller(game *Game) {
	game.Events <- Event{
		GameID:    game.ID,
		Type:      Start,
		Msg:       game.Process.Teams[game.Process.Storyteller],
		Receivers: game.PlayersIDs,
	}
}

func NotifyWord(game *Game, story string) {
	game.Events <- Event{
		GameID:    game.ID,
		Receivers: []uint{game.Process.Teams[game.Process.Storyteller]},
		Type:      Story,
		Msg:       story,
	}
}

func (g *Game) MakeTurn(id uint) {
	story, found := g.nextWord()

	if !found {
		NotifyGameEnded(g)
		return
	}

	NotifyWord(g, story)

	timer := time.NewTicker(1 * time.Second)
	timerDone := make(chan struct{})
	go tick(g, timerDone, timer)
	go func(timerDone chan struct{}) {
		time.Sleep(time.Duration(g.Timer) * time.Second)
		timer.Stop()
		close(timerDone)
	}(timerDone)
}

func tick(game *Game, timerDone chan struct{}, timer *time.Ticker) {
	i := game.Timer
	defer fmt.Printf("Closing tick\n")

	for {
		select {
		case <-timerDone:
			game.Process.Storyteller = (game.Process.Storyteller + 1) % game.NumPlayers
			NotifyStoryteller(game)
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
			game.Events <- Event{
				GameID:    game.ID,
				Type:      Tick,
				Msg:       i,
				Receivers: game.PlayersIDs,
			}
		}
	}
}
