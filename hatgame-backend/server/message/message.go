package message

type Type string

const (
	Tick     Type = "tick"
	Ready    Type = "ready"
	AddWord  Type = "word"
	Guess    Type = "guess"
	GameInfo Type = "game"
	Team     Type = "team"
	End      Type = "end"
	Start    Type = "start"
	Story    Type = "story"
	Error    Type = "error"
)

type Message struct {
	Type Type
	Msg  interface{}
}
