package containers

import (
	"encoding/json"
	"fmt"

	"github.com/gorilla/websocket"
)

type Message struct {
	Type string
	Msg  interface{}
}

func CreateMessage(t string, m interface{}) ([]byte, error) {
	return json.Marshal(Message{Type: t, Msg: m})
}

func (msg Message) HandleMessage(
	ws *websocket.Conn,
	game *Game,
	id uint,
	timerGameEnd chan struct{},
	errors chan error) {
	fmt.Printf("HandleMessage: %s\n", msg)
	switch msg.Type {
	case "word":
		word := fmt.Sprintf("%s", msg.Msg)
		resp, err := game.AddWord(id, word)
		if err != nil {
			errors <- err
			return
		}
		err = ws.WriteMessage(websocket.TextMessage, resp)
		if err != nil {
			errors <- fmt.Errorf("could not send message")
			return
		}

		if game.CheckWordsFinished() {
			err := Start(id, game)
			if err != nil {
				errors <- err
				return
			}
		}
	case "ready":
		fmt.Printf("Storyteller %d is ready\n", id)
		err := MakeTurn(id, game, timerGameEnd)
		if err != nil {
			errors <- err
			return
		}
	case "guess":
		word := fmt.Sprintf("%s", msg.Msg)
		game.Process.guessWord(word)
		fmt.Printf("Guessed word %s\n", word)
		ok, err := PickWord(game)
		if !ok {
			timerGameEnd <- struct{}{}
		}
		if err != nil {
			errors <- err
			return
		}
	default:
		errors <- fmt.Errorf("can't decode message: %s", msg)
	}
}
