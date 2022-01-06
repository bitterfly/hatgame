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
	switch msg.Type {
	case "word":
		word := fmt.Sprintf("Word: %s", msg.Msg)
		resp, err := game.AddWord(id, word)
		if err != nil {
			return err
		}
		err = ws.WriteMessage(websocket.TextMessage, resp)
		if err != nil {
			return fmt.Errorf("could not send message")
		}
	default:
		fmt.Printf("Type: %s\n", msg.Type)
	}
	return nil
}
