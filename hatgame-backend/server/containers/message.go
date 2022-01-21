package containers

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/gorilla/websocket"
)

type Message struct {
	Type string
	Msg  interface{}
}

func (m Message) String() string {
	switch m.Type {
	case "ready":
		return m.Type
	default:
		return fmt.Sprintf("%s %s", m.Type, m.Msg)
	}
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
	log.Printf("HandleMessage: %s\n", msg)
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
			err := game.Start(id)
			if err != nil {
				errors <- err
				return
			}
		}
	case "ready":
		err := MakeTurn(id, game, timerGameEnd)
		if err != nil {
			errors <- err
			return
		}
	case "guess":
		word := fmt.Sprintf("%s", msg.Msg)
		game.Process.guessWord(word)

		story, found := game.nextWord()

		if !found {
			err := NotifyGameEnded(game)
			if err != nil {
				errors <- err
				return
			}
			timerGameEnd <- struct{}{}
			for i := 0; i < game.NumPlayers; i++ {
				game.Process.GameEnd <- struct{}{}
			}
			close(game.Process.GameEnd)
			return
		}

		err := NotifyWord(game, story)
		if err != nil {
			errors <- err
			return
		}
	default:
		errors <- fmt.Errorf("can't decode message: %s", msg)
	}
}
