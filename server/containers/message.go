package containers

import (
	"fmt"

	"github.com/gorilla/websocket"
)

type Message struct {
	Type string
	Msg  interface{}
}

func (msg Message) HandleMessage(ws *websocket.Conn, game *Game, id uint, wordGuessed chan struct{}) error {
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
		game.Process.TimerLeft = game.Timer
		fmt.Printf("Timer is %d\n", game.Process.TimerLeft)
		MakeTurn(id, game, wordGuessed)
	case "guess":
		// TODO: maybe detach the timer stopping and the next word
		word := fmt.Sprintf("%s", msg.Msg)
		wordGuessed <- struct{}{}
		fmt.Printf("Guessed word %s\n", word)
		game.Process.guessWord(word)
		fmt.Printf("Timer is %d\n", game.Process.TimerLeft)
		MakeTurn(id, game, wordGuessed)
	default:
		fmt.Printf("Type: %s\n", msg.Type)
	}
	return nil
}
