package game

import (
	"fmt"
	"reflect"
	"sort"
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
		t.Errorf("Adding player should send 1 event instead of %d", len(events))
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

func ToString(receivers map[uint]struct{}) string {
	res := make([]uint, 0, len(receivers))
	for r := range receivers {
		res = append(res, r)
	}
	sort.Slice(res, func(i, j int) bool { return res[i] < res[j] })
	return fmt.Sprintf("%v", res)
}

func (event Event) String() string {

	return fmt.Sprintf("GameID: %d\nMsg: %s\nType: %s\nReceivers: %v",
		event.GameID,
		event.Msg,
		event.Type,
		ToString(event.Receivers),
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
			ToString(game.Players.IDs))
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
			ToString(game.Players.IDs))
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
			ToString(game.Players.IDs))
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
