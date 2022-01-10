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

func (msg Message) HandleMessage(ws *websocket.Conn, game *Game, id uint, timerGameEnd chan struct{}) error {
	fmt.Printf("HandleMessage: %s\n", msg)
	switch msg.Type {
	case "word":
		word := fmt.Sprintf("%s", msg.Msg)
		resp, err := game.AddWord(id, word)
		if err != nil {
			return err
		}
		err = ws.WriteMessage(websocket.TextMessage, resp)
		if err != nil {
			return fmt.Errorf("could not send message")
		}

		if game.CheckWordsFinished() {
			Start(id, game)
		}
	case "ready":
		fmt.Printf("Storyteller %d is ready\n", id)
		err := MakeTurn(id, game, timerGameEnd)
		if err != nil {
			fmt.Printf("%s\n", err)
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
			fmt.Printf("%s\n", err)
		}
	default:
		fmt.Printf("Type: %s\n", msg.Type)
	}
	return nil
}
