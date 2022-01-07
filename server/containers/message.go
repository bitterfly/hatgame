package containers

import (
	"fmt"

	"github.com/gorilla/websocket"
)

type Message struct {
	Type string
	Msg  interface{}
}

func (msg Message) HandleMessage(ws *websocket.Conn, game *Game, id uint) error {
	timerDone := make(chan struct{})
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
		MakeTurn(id, game, timerDone)
	default:
		fmt.Printf("Type: %s\n", msg.Type)
	}
	return nil
}
