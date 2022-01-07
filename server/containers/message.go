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
	fmt.Printf("HandleMessage: %s, type: %s, msg: %s\n", msg, msg.Type, msg.Msg)
	switch msg.Type {
	case "word":
		fmt.Printf("case word\n")
		word := fmt.Sprintf("%s", msg.Msg)
		fmt.Printf("Adding word\n")
		resp, err := game.AddWord(id, word)
		if err != nil {
			return err
		}
		fmt.Printf("Writing message: %s\n", resp)
		err = ws.WriteMessage(websocket.TextMessage, resp)
		if err != nil {
			return fmt.Errorf("could not send message")
		}

		if game.CheckWordsFinished() {
			Start(id, game)
		}
	case "ready":
		fmt.Printf("Storyteller %d is ready", id)
	default:
		fmt.Printf("Type: %s\n", msg.Type)
	}
	return nil
}
