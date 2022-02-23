package containers

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/gorilla/websocket"
)

type MessageType string

const (
	Tick     MessageType = "tick"
	Ready    MessageType = "ready"
	AddWord  MessageType = "word"
	Guess    MessageType = "guess"
	GameInfo MessageType = "game"
	Team     MessageType = "team"
	End      MessageType = "end"
	Start    MessageType = "start"
	Story    MessageType = "story"
)

type Message struct {
	Type MessageType
	Msg  interface{}
}

func (m Message) String() string {
	switch m.Type {
	case Ready:
		return "ready"
	default:
		return fmt.Sprintf("%s %s", m.Type, m.Msg)
	}
}

func CreateMessage(t MessageType, m interface{}) ([]byte, error) {
	return json.Marshal(Message{Type: t, Msg: m})
}

func (msg Message) HandleMessage(
	ws *websocket.Conn,
	game *Game,
	id uint,
	errors chan error) {
	log.Printf("HandleMessage: %s\n", msg)
	switch msg.Type {
	case AddWord:
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
			game.Start()
		}
	case Ready:
		err := MakeTurn(id, game)
		if err != nil {
			errors <- err
			return
		}
	case Guess:
		word := fmt.Sprintf("%s", msg.Msg)
		game.Process.guessWord(word)

		// story, found := game.nextWord()

		// if !found {
		// err := NotifyGameEnded(game)
		// if err != nil {
		// 	errors <- err
		// 	return
		// }
		// close(game.Process.GameEnd)
		// return
		// }

		// err := NotifyWord(game, story)
		// if err != nil {
		// 	errors <- err
		// 	return
		// }
	default:
		errors <- fmt.Errorf("can't decode message: %s", msg)
	}
}
