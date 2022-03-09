package game

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"sort"
	"sync"
	"time"

	"github.com/bitterfly/go-chaos/hatgame/server/containers"
	"github.com/bitterfly/go-chaos/hatgame/utils"
)

type EventType string

const (
	EventRequestToStart   EventType = "request_to_start"
	EventReadyToStart     EventType = "ready_to_start"
	EventWordPhaseStart   EventType = "word_phase_start"
	EventGameInfo         EventType = "game"
	EventTick             EventType = "tick"
	EventTeam             EventType = "team"
	EventEnd              EventType = "end"
	EventGuessPhaseStart  EventType = "guess_phase_start"
	EventStory            EventType = "story"
	EventError            EventType = "error"
	EventAddWord          EventType = "add_word"
	EventReadyStoryteller EventType = "ready_storyteller"
	EventGuess            EventType = "guess"
	EventQuitLobby        EventType = "quit_lobby"
)

type Event struct {
	Type      EventType
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
}

type Players struct {
	IDs   map[uint]struct{}
	Users map[uint]containers.User
	Quit  map[uint]chan struct{}
}

func (p Players) MarshalJSON() ([]byte, error) {
	Players := make([]containers.User, 0, len(p.Users))
	for _, v := range p.Users {
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

func (g *Game) CloseChannels() {
}

func (g *Game) GetNextWord() {
	word, found := g.nextWord()
	if found {
		NotifyWord(g, word)
		return
	}
	NotifyGameEnded(g)
	close(g.Process.GameEnd)
	close(g.Events)
}

func (g *Game) RemovePlayer(id uint) {
	if g.Host == id {
		for _, quitChan := range g.Players.Quit {
			quitChan <- struct{}{}
		}
		g.Events <- Event{
			GameID:    g.ID,
			Type:      EventQuitLobby,
			Receivers: g.Players.IDs,
		}
		close(g.Events)
		return
	}
	g.Players.Quit[id] <- struct{}{}
	close(g.Players.Quit[id])
	delete(g.Players.Quit, id)
	delete(g.Players.IDs, id)
	delete(g.Players.Users, id)
	g.Events <- Event{
		GameID:    g.ID,
		Type:      EventQuitLobby,
		Receivers: map[uint]struct{}{id: {}},
	}
	g.Events <- Event{
		GameID:    g.ID,
		Type:      EventGameInfo,
		Msg:       g,
		Receivers: g.Players.IDs,
	}
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
		Players: Players{
			IDs:   map[uint]struct{}{host.ID: {}},
			Users: map[uint]containers.User{host.ID: host},
			Quit:  map[uint]chan struct{}{host.ID: make(chan struct{})},
		},
	}
}

func (g *Game) AddPlayer(user containers.User) bool {
	if len(g.Players.IDs) == g.NumPlayers {
		g.Events <- Event{
			GameID:    g.ID,
			Type:      EventError,
			Msg:       "too many players",
			Receivers: map[uint]struct{}{user.ID: {}},
		}
		return false
	}
	if _, ok := g.Players.IDs[user.ID]; ok {
		g.Events <- Event{
			GameID:    g.ID,
			Type:      EventError,
			Msg:       "player already in game",
			Receivers: map[uint]struct{}{user.ID: {}},
		}
		return false
	}
	g.Words.Mutex.Lock()
	defer g.Words.Mutex.Unlock()
	g.Words.ByUser[user.ID] = make(map[string]struct{})
	g.Players.IDs[user.ID] = struct{}{}
	g.Players.Quit[user.ID] = make(chan struct{})
	g.Players.Users[user.ID] = user
	return true
}

func (g *Game) addWord(id uint, word string) error {
	g.Words.Mutex.Lock()
	defer g.Words.Mutex.Unlock()
	if _, ok := g.Players.IDs[id]; !ok {
		return fmt.Errorf("no player with id %d", id)
	}
	if len(g.Words.ByUser[id]) == g.NumWords {
		return fmt.Errorf("words limit reached")
	}
	if _, ok := g.Words.All[word]; ok {
		return fmt.Errorf("already used this word")
	}
	g.Words.ByUser[id][word] = struct{}{}
	g.Words.All[word] = struct{}{}
	return nil
}

func (g *Game) AddWord(id uint, word string) {
	err := g.addWord(id, word)

	if err != nil {
		g.Events <- Event{
			GameID:    g.ID,
			Type:      EventError,
			Msg:       err.Error(),
			Receivers: map[uint]struct{}{id: {}},
		}
		return
	}

	g.Events <- Event{
		GameID:    g.ID,
		Type:      EventAddWord,
		Msg:       word,
		Receivers: map[uint]struct{}{id: {}},
	}

	if g.CheckWordsFinished() {
		g.MakeTeams()
		NotifyGuessPhaseStart(g)
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

func (g *Game) StartWordPhase() {
	g.Events <- Event{
		GameID:    g.ID,
		Type:      EventWordPhaseStart,
		Receivers: g.Players.IDs,
	}
}

func NotifyGuessPhaseStart(g *Game) {
	for i, id := range g.Process.Teams {
		g.Events <- Event{
			GameID:    g.ID,
			Type:      EventTeam,
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
		Type:      EventEnd,
		Msg:       game.Process.Result,
	}
}

func NotifyStoryteller(game *Game) {
	game.Events <- Event{
		GameID:    game.ID,
		Type:      EventGuessPhaseStart,
		Msg:       game.Process.Teams[game.Process.Storyteller],
		Receivers: game.Players.IDs,
	}
}

func NotifyWord(game *Game, story string) {
	game.Events <- Event{
		GameID: game.ID,
		Receivers: map[uint]struct{}{
			game.Process.Teams[game.Process.Storyteller]: {}},
		Type: EventStory,
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
	defer timer.Stop()
	go tick(g, timer)
	for {
		select {
		case <-time.After(time.Duration(g.Timer) * time.Second):
			fmt.Println("Timer out")
			g.Process.Storyteller = (g.Process.Storyteller + 1) % g.NumPlayers
			NotifyStoryteller(g)
			return
		case _, ok := <-g.Process.GameEnd:
			if !ok {
				return
			}
		}
	}
}

func tick(game *Game, timer *time.Ticker) {
	i := game.Timer
	for _ = range timer.C {
		fmt.Println("tick")
		i -= 1
		game.Events <- Event{
			GameID:    game.ID,
			Type:      EventTick,
			Msg:       i,
			Receivers: game.Players.IDs,
		}
	}
}
