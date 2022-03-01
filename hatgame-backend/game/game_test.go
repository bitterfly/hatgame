package game

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
	"sync"
	"testing"

	"github.com/bitterfly/go-chaos/hatgame/server/containers"
)

func checkEvents(t *testing.T, wg *sync.WaitGroup, game *Game,
	expectedNum int, expectedEvent Event) {
	defer wg.Done()
	events := make([]Event, 0)
	for e := range game.Events {
		events = append(events, e)
	}
	if len(events) != expectedNum {
		t.Errorf("Function should send %d event instead of %d\nEvents: %v",
			expectedNum,
			len(events),
			events)
		return
	}
	if len(events) == 0 {
		return
	}
	if !compareEvents(events[0], expectedEvent) {
		t.Errorf("The received event:\n%s\n does not match the expected event:\n%s",
			events[0],
			expectedEvent)
	}
}

func skipEvents(wg *sync.WaitGroup, game *Game) {
	defer wg.Done()
	for range game.Events {
	}
}

func compareEvents(first, second Event) bool {
	return first.GameID == second.GameID &&
		first.Msg == second.Msg &&
		first.Type == second.Type &&
		reflect.DeepEqual(first.Receivers, second.Receivers)
}

func PrintUintSet(receivers map[uint]struct{}) string {
	res := make([]uint, 0, len(receivers))
	for r := range receivers {
		res = append(res, r)
	}
	sort.Slice(res, func(i, j int) bool { return res[i] < res[j] })
	return fmt.Sprintf("%v", res)
}

func PrintStringSet(words map[string]struct{}) string {
	res := make([]string, 0, len(words))
	for r := range words {
		res = append(res, r)
	}
	sort.Strings(res)
	return fmt.Sprintf("%v", res)
}

func PrintUintStringSet(words map[uint]map[string]struct{}) string {
	keys := make([]uint, 0, len(words))
	for r := range words {
		keys = append(keys, r)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	res := make([]string, 0, len(keys))
	for _, k := range keys {
		res = append(res, fmt.Sprintf("%d: %s", k, PrintStringSet(words[k])))
	}
	return fmt.Sprintf("{%v}", strings.Join(res, ","))
}

func (event Event) String() string {

	return fmt.Sprintf("GameID: %d\nMsg: %s\nType: %s\nReceivers: %v",
		event.GameID,
		event.Msg,
		event.Type,
		PrintUintSet(event.Receivers),
	)
}

func TestAddPlayer_AddExistingUserReturn(t *testing.T) {
	users := []containers.User{
		{
			ID:       1,
			Email:    "1",
			Username: "1"},
	}
	game := NewGame(1, users[0], 2, 1, 1)
	var wg sync.WaitGroup
	defer wg.Wait()
	wg.Add(1)
	go skipEvents(&wg, game)
	res := game.AddPlayer(users[0])
	close(game.Events)
	if res {
		t.Errorf("should not be able to add player with id %d in game with players: %v",
			users[0].ID,
			PrintUintSet(game.Players.IDs))
	}
}

func TestAddPlayer_AddExistingUserEvent(t *testing.T) {
	users := []containers.User{
		{
			ID:       1,
			Email:    "1",
			Username: "1"}}
	game := NewGame(1, users[0], 2, 1, 1)
	var wg sync.WaitGroup
	defer wg.Wait()
	wg.Add(1)
	expected := Event{
		GameID:    game.ID,
		Type:      EventError,
		Msg:       "player already in game",
		Receivers: map[uint]struct{}{users[0].ID: {}}}
	go checkEvents(t, &wg, game, 1, expected)
	game.AddPlayer(users[0])
	close(game.Events)
}

func TestAddPlayer_HitLimitReturn(t *testing.T) {
	users := []containers.User{
		{
			ID:       1,
			Email:    "1",
			Username: "1"},
		{
			ID:       2,
			Email:    "2",
			Username: "2"},
		{
			ID:       3,
			Email:    "3",
			Username: "3"}}
	game := NewGame(1, users[0], 2, 1, 1)
	var wg sync.WaitGroup
	defer wg.Wait()
	wg.Add(1)
	go skipEvents(&wg, game)
	game.AddPlayer(users[1])
	res := game.AddPlayer(users[2])
	close(game.Events)
	if res {
		t.Errorf(
			"should not be able to add player in game with cap %d and joined players: %v",
			game.NumPlayers,
			PrintUintSet(game.Players.IDs))
	}
}

func TestAddPlayer_HitLimitEvent(t *testing.T) {
	users := []containers.User{
		{
			ID:       1,
			Email:    "1",
			Username: "1"},
		{
			ID:       2,
			Email:    "2",
			Username: "2"},
		{
			ID:       3,
			Email:    "3",
			Username: "3"}}
	game := NewGame(1, users[0], 2, 1, 1)
	var wg sync.WaitGroup
	defer wg.Wait()
	wg.Add(1)
	expected := Event{
		GameID:    game.ID,
		Type:      EventError,
		Msg:       "too many players",
		Receivers: map[uint]struct{}{users[2].ID: {}}}
	go checkEvents(t, &wg, game, 1, expected)
	game.AddPlayer(users[1])
	game.AddPlayer(users[2])
	close(game.Events)
}

func TestAddPlayer_SuccessResult(t *testing.T) {
	users := []containers.User{
		{
			ID:       1,
			Email:    "1",
			Username: "1"},
		{
			ID:       2,
			Email:    "2",
			Username: "2"},
	}
	game := NewGame(1, users[0], 2, 1, 1)
	var wg sync.WaitGroup
	defer wg.Wait()
	wg.Add(1)
	go skipEvents(&wg, game)
	res := game.AddPlayer(users[1])
	close(game.Events)
	if !res {
		t.Errorf("should be able to add player with id %d in game with players: %v",
			users[0].ID,
			PrintUintSet(game.Players.IDs))
	}
}

func TestAddPlayer_SuccessEvent(t *testing.T) {
	users := []containers.User{
		{
			ID:       1,
			Email:    "1",
			Username: "1"},
		{
			ID:       2,
			Email:    "2",
			Username: "2"},
	}
	game := NewGame(1, users[0], 2, 1, 1)
	var wg sync.WaitGroup
	defer wg.Wait()
	wg.Add(1)
	go checkEvents(t, &wg, game, 0, Event{})
	game.AddPlayer(users[1])
	close(game.Events)
}

func TestAddWord_NoPlayerWithID(t *testing.T) {
	users := []containers.User{
		{
			ID:       1,
			Email:    "1",
			Username: "1"},
	}
	game := NewGame(1, users[0], 2, 1, 1)
	var wg sync.WaitGroup
	defer wg.Wait()
	wg.Add(1)
	var ID uint = 2
	expected := Event{
		GameID:    game.ID,
		Type:      EventError,
		Msg:       fmt.Sprintf("no player with id %d", ID),
		Receivers: map[uint]struct{}{ID: {}},
	}
	go checkEvents(t, &wg, game, 1, expected)
	game.AddWord(ID, "foo")
	close(game.Events)
}

func TestAddWord_SuccessEmpty(t *testing.T) {
	users := []containers.User{
		{
			ID:       1,
			Email:    "1",
			Username: "1"},
	}
	game := NewGame(1, users[0], 2, 1, 1)
	var wg sync.WaitGroup
	defer wg.Wait()
	wg.Add(1)
	word := "foo"
	expectedEvent := Event{
		GameID:    game.ID,
		Type:      EventAddWord,
		Msg:       word,
		Receivers: map[uint]struct{}{users[0].ID: {}},
	}
	go checkEvents(t, &wg, game, 1, expectedEvent)
	game.AddWord(users[0].ID, word)
	close(game.Events)
	expectedAll := map[string]struct{}{word: {}}
	if !reflect.DeepEqual(game.Words.All, expectedAll) {
		t.Errorf("After adding the word %s all words in game should be %v instead of %v",
			word,
			PrintStringSet(expectedAll),
			PrintStringSet(game.Words.All),
		)
	}
	expectedByUser := map[uint]map[string]struct{}{
		users[0].ID: {word: {}}}
	if !reflect.DeepEqual(game.Words.ByUser, expectedByUser) {
		t.Errorf("After adding the word %s all words in game should be %v instead of %v",
			word,
			PrintUintStringSet(expectedByUser),
			PrintUintStringSet(game.Words.ByUser),
		)
	}
}

func TestAddWord_SuccessFull(t *testing.T) {
	users := []containers.User{
		{
			ID:       1,
			Email:    "1",
			Username: "1"},

		{
			ID:       2,
			Email:    "2",
			Username: "2"},
	}
	words := []string{
		"foo",
		"bar",
	}
	game := NewGame(1, users[0], 2, 2, 1)
	var wg sync.WaitGroup
	defer wg.Wait()
	expectedEvent := Event{
		GameID:    game.ID,
		Type:      EventAddWord,
		Msg:       words[1],
		Receivers: map[uint]struct{}{users[1].ID: {}},
	}
	wg.Add(1)
	go checkEvents(t, &wg, game, 1, expectedEvent)
	game.AddPlayer(users[1])
	game.Words.ByUser[users[0].ID][words[0]] = struct{}{}
	game.Words.All[words[0]] = struct{}{}
	game.AddWord(users[1].ID, words[1])
	close(game.Events)
	expectedAll := map[string]struct{}{words[0]: {}, words[1]: {}}
	if !reflect.DeepEqual(game.Words.All, expectedAll) {
		t.Errorf("After adding the word %s all words in game should be %v instead of %v",
			words[1],
			PrintStringSet(expectedAll),
			PrintStringSet(game.Words.All),
		)
	}
	expectedByUser := map[uint]map[string]struct{}{
		users[0].ID: {words[0]: {}},
		users[1].ID: {words[1]: {}},
	}
	if !reflect.DeepEqual(game.Words.ByUser, expectedByUser) {
		t.Errorf("After adding the word %s all words in game should be %v instead of %v",
			words[1],
			PrintUintStringSet(expectedByUser),
			PrintUintStringSet(game.Words.ByUser),
		)
	}
}
