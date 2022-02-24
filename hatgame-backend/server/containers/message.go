package containers

import (
	"encoding/json"
	"fmt"
	"log"
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
	Error    MessageType = "error"
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
	game *Game,
	id uint) {
	log.Printf("HandleMessage: %s\n", msg)
	switch msg.Type {
	case AddWord:
		word := fmt.Sprintf("%s", msg.Msg)
		game.AddWord(id, word)
	case Ready:
		game.MakeTurn(id)
	case Guess:
		word := fmt.Sprintf("%s", msg.Msg)
		game.Process.guessWord(word)
		game.processNextWord()

	default:
		game.Errors <- fmt.Errorf("can't decode message: %s", msg)
	}
}
