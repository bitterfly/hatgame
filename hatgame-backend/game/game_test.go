package game

import (
	"fmt"
	"reflect"
	"sort"
	"sync"
	"testing"

	"github.com/bitterfly/go-chaos/hatgame/server/containers"
)

func compareEvents(first, second Event) bool {
	return first.GameID == second.GameID &&
		first.Msg == second.Msg &&
		first.Type == second.Type &&
		reflect.DeepEqual(first.Receivers, second.Receivers)
}

func (event Event) String() string {
	receivers := make([]uint, len(event.Receivers))
	for r := range event.Receivers {
		receivers = append(receivers, r)
	}
	sort.Slice(receivers, func(i, j int) bool { return receivers[i] < receivers[j] })

	return fmt.Sprintf("GameID: %d\nMsg: %s\nType: %s\nReceivers: %v",
		event.GameID,
		event.Msg,
		event.Type,
		receivers,
	)
}

func TestAddPlayer_AddExistingUserReturn(t *testing.T) {
	host := containers.User{
		ID:       1,
		Email:    "1",
		Username: "1"}
	game := NewGame(1, host, 2, 1, 1)
	var wg sync.WaitGroup
	defer wg.Wait()
	wg.Add(1)
	go func() {
		defer wg.Done()
		for range game.Events {
		}
	}()

	res := game.AddPlayer(host)
	close(game.Events)

	if res {
		t.Errorf("should not be able to add player with id %d in game with players: %v", host.ID, game.Players.IDs)
	}
}

func TestAddPlayer_AddExistingUserEvent(t *testing.T) {
	host := containers.User{
		ID:       1,
		Email:    "1",
		Username: "1"}
	game := NewGame(1, host, 2, 1, 1)
	var wg sync.WaitGroup
	defer wg.Wait()
	wg.Add(1)
	go func() {
		defer wg.Done()
		events := make([]Event, 0)
		for e := range game.Events {
			events = append(events, e)
		}
		if len(events) != 1 {
			t.Errorf("Adding player should send 1 event instead of %d", len(events))
			return
		}
		expected := Event{
			GameID:    game.ID,
			Type:      EventError,
			Msg:       "player already in game",
			Receivers: map[uint]struct{}{host.ID: {}}}
		if !compareEvents(events[0], expected) {
			t.Errorf("The received event:\n%s\n does not match the expected event:\n%s",
				events[0],
				expected)
		}
	}()

	game.AddPlayer(host)
	close(game.Events)
}
