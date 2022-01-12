package containers

type Statistics struct {
	GamesPlayed  int64
	NumberOfWins int64
	TopWords     []Word
}

type Word struct {
	Word  string
	Count int
}
