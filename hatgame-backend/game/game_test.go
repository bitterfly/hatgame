package game

import (
	"fmt"
	"reflect"
	"sort"
	"sync"
	"testing"

	"github.com/bitterfly/go-chaos/hatgame/server/containers"
)

func checkEvents(t *testing.T, wg *sync.WaitGroup, game *Game, expected Event) {
	defer wg.Done()
	events := make([]Event, 0)
	for e := range game.Events {
		events = append(events, e)
	}
	if len(events) != 1 {
		t.Errorf("Adding player should send 1 event instead of %d", len(events))
		return
	}
	if !compareEvents(events[0], expected) {
		t.Errorf("The received event:\n%s\n does not match the expected event:\n%s",
			events[0],
			expected)
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
	go checkEvents(t, &wg, game, expected)
	game.AddPlayer(users[0])
	close(game.Events)
}
