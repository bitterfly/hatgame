package game

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"sort"
	"sync"
	"time"

	"github.com/bitterfly/go-chaos/hatgame/server/containers"
	"github.com/bitterfly/go-chaos/hatgame/server/message"
	"github.com/bitterfly/go-chaos/hatgame/utils"
)

type Event struct {
	Type      message.Type
	Msg       interface{}
	Receivers map[uint]struct{}
	GameID    uint
}

type Game struct {
	ID         uint
	Host       uint
	NumPlayers int
	Timer      int
	NumWords   int
	Players    Players
	Words      Words      `json:"-"`
	Process    Process    `json:"-"`
	Events     chan Event `json:"-"`
	Errors     chan error `json:"-"`
}

type Players struct {
	IDs   map[uint]struct{}
	Users map[containers.User]struct{}
}

func (p Players) MarshalJSON() ([]byte, error) {
	Players := make([]containers.User, 0, len(p.Users))
	for v := range p.Users {
		Players = append(Players, v)
	}
	return json.Marshal(Players)
}

type Words struct {
	ByUser map[uint]map[string]struct{}
	All    map[string]struct{}
	Mutex  *sync.RWMutex
}

type Process struct {
	WordID       int
	Storyteller  int
	Teams        []uint
	Result       []containers.Result
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
	g.Process.Result = make([]containers.Result, 0, teams)

	for i := 0; i < teams; i++ {
		first, second := utils.Order(
			g.Process.Teams[i],
			g.Process.Teams[(i+teams)%len(g.Process.Teams)])

		res := containers.Result{FirstID: first, SecondID: second}
		res.Score = rev[res.FirstID] + rev[res.SecondID]
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
	g.Words.Mutex.RLock()
	defer g.Words.Mutex.RUnlock()

	if len(g.Words.All) == len(g.Process.GuessedWords) {
		return "", false
	}

	unguessed := make([]string, 0, len(g.Words.All)-len(g.Process.GuessedWords))
	for word := range g.Words.All {
		if _, ok := g.Process.GuessedWords[word]; !ok {
			unguessed = append(unguessed, word)
		}
	}

	return unguessed[rand.Intn(len(unguessed))], true
}

func NewGame(gameID uint, host containers.User, numPlayers, numWords, timer int) *Game {
	wordsByUser := make(map[uint]map[string]struct{})
	wordsByUser[host.ID] = make(map[string]struct{})
	words := make(map[string]struct{})
	rand.Seed(time.Now().UnixNano())

	return &Game{
		ID: gameID,
		Words: Words{
			ByUser: wordsByUser,
			All:    words,
			Mutex:  &sync.RWMutex{},
		},
		Process: Process{
			Teams:        make([]uint, 0, numPlayers),
			GuessedWords: make(map[string]uint),
			Mutex:        &sync.RWMutex{},
			GameEnd:      make(chan struct{}),
			Storyteller:  0,
			WordID:       0,
		},
		NumPlayers: numPlayers,
		NumWords:   numWords,
		Timer:      timer,
		Host:       host.ID,
		Events:     make(chan Event),
		Errors:     make(chan error),
		Players: Players{
			IDs:   map[uint]struct{}{host.ID: {}},
			Users: map[containers.User]struct{}{host: {}},
		},
	}
}

func (g *Game) AddPlayer(user containers.User) bool {
	if len(g.Players.IDs) == g.NumPlayers {
		g.Events <- Event{
			GameID:    g.ID,
			Type:      message.Error,
			Msg:       "too many players",
			Receivers: map[uint]struct{}{user.ID: {}},
		}
		return false
	}
	if _, ok := g.Players.IDs[user.ID]; ok {
		g.Events <- Event{
			GameID:    g.ID,
			Type:      message.Error,
			Msg:       "player already in game",
			Receivers: map[uint]struct{}{user.ID: {}},
		}
		return false
	}
	g.Words.Mutex.Lock()
	defer g.Words.Mutex.Unlock()
	g.Words.ByUser[user.ID] = make(map[string]struct{})
	g.Players.IDs[user.ID] = struct{}{}
	g.Players.Users[user] = struct{}{}
	return true
}

func (g *Game) addWord(id uint, word string) (bool, error) {
	g.Words.Mutex.Lock()
	defer g.Words.Mutex.Unlock()
	if _, ok := g.Players.IDs[id]; !ok {
		return false, fmt.Errorf("no player with id %d", id)
	}
	if len(g.Words.ByUser[id]) == g.NumWords {
		return false, fmt.Errorf("words limit reached")
	}
	if _, ok := g.Words.All[word]; ok {
		return false, nil
	}
	g.Words.ByUser[id][word] = struct{}{}
	g.Words.All[word] = struct{}{}
	return true, nil
}

func (g *Game) AddWord(id uint, word string) {
	ok, err := g.addWord(id, word)
	if err != nil {
		g.Errors <- err
	}

	if !ok {
		g.Events <- Event{
			GameID:    g.ID,
			Type:      message.Error,
			Msg:       "Already used this word",
			Receivers: map[uint]struct{}{id: {}},
		}
		return
	}

	g.Events <- Event{
		GameID:    g.ID,
		Type:      message.AddWord,
		Msg:       word,
		Receivers: map[uint]struct{}{id: {}},
	}

	if g.CheckWordsFinished() {
		g.MakeTeams()
		NotifyGameStarted(g)
		NotifyStoryteller(g)
	}
}

func (g *Game) CheckWordsFinished() bool {
	g.Words.Mutex.RLock()
	defer g.Words.Mutex.RUnlock()
	return len(g.Words.All) == g.NumPlayers*g.NumWords
}

func (g *Game) MakeTeams() {
	g.Words.Mutex.RLock()
	for id := range g.Words.ByUser {
		g.Process.Teams = append(g.Process.Teams, id)
	}
	g.Words.Mutex.RUnlock()

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
			Type:      message.Team,
			Msg:       g.Process.Teams[(i+int(float64(g.NumPlayers)/2))%g.NumPlayers],
			Receivers: map[uint]struct{}{id: {}},
		}
	}
}

func NotifyGameEnded(game *Game) {
	game.GetResults()
	game.Events <- Event{
		GameID:    game.ID,
		Receivers: game.Players.IDs,
		Type:      message.End,
		Msg:       game.Process.Result,
	}
}

func NotifyStoryteller(game *Game) {
	game.Events <- Event{
		GameID:    game.ID,
		Type:      message.Start,
		Msg:       game.Process.Teams[game.Process.Storyteller],
		Receivers: game.Players.IDs,
	}
}

func NotifyWord(game *Game, story string) {
	game.Events <- Event{
		GameID: game.ID,
		Receivers: map[uint]struct{}{
			game.Process.Teams[game.Process.Storyteller]: {}},
		Type: message.Story,
		Msg:  story,
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
				Type:      message.Tick,
				Msg:       i,
				Receivers: game.Players.IDs,
			}
		}
	}
}
