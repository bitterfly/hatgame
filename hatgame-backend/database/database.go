package database

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/bitterfly/go-chaos/hatgame/game"
	"github.com/bitterfly/go-chaos/hatgame/schema"
	"github.com/bitterfly/go-chaos/hatgame/server/containers"
	"github.com/bitterfly/go-chaos/hatgame/utils"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/exp/rand"
	"gonum.org/v1/gonum/stat/sampleuv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type psqlInfo struct {
	Host     string
	Port     int
	User     string
	Password string
	Dbname   string
	Sslmode  string
}

type ErrorType int

const (
	InsertError ErrorType = iota
	ConflictError
	OpenError
	ConfigError
	MigrateError
	UpdateError
	QueryError
)

type DatabaseError struct {
	ErrorType ErrorType
	msg       error
}

func newMigrateError(err error) *DatabaseError {
	if err == nil {
		return nil
	}
	return &DatabaseError{
		ErrorType: MigrateError,
		msg:       fmt.Errorf("database migrate error: %w", err),
	}
}

func newConflictError(err error) *DatabaseError {
	if err == nil {
		return nil
	}
	return &DatabaseError{
		ErrorType: ConflictError,
		msg:       fmt.Errorf("database create error: %w", err),
	}
}

func newOpenError(err error) *DatabaseError {
	if err == nil {
		return nil
	}
	return &DatabaseError{
		ErrorType: OpenError,
		msg:       fmt.Errorf("database open error: %w", err),
	}
}

func newInsertError(err error) *DatabaseError {
	if err == nil {
		return nil
	}
	return &DatabaseError{
		ErrorType: InsertError,
		msg:       fmt.Errorf("database insert error: %w", err),
	}
}

func newConfigErrror(err error) *DatabaseError {
	if err == nil {
		return nil
	}
	return &DatabaseError{
		ErrorType: ConfigError,
		msg:       fmt.Errorf("database config error: %w", err),
	}
}

func newUpdateError(err error) *DatabaseError {
	if err == nil {
		return nil
	}
	return &DatabaseError{
		ErrorType: UpdateError,
		msg:       fmt.Errorf("database update error: %w", err),
	}
}

func newQueryError(err error) *DatabaseError {
	if err == nil {
		return nil
	}
	return &DatabaseError{
		ErrorType: QueryError,
		msg:       fmt.Errorf("database query error: %w", err),
	}
}

func (e *DatabaseError) Error() string {
	return e.msg.Error()
}

func (p psqlInfo) String() string {
	return fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=%s",
		p.Host, p.Port, p.User, p.Password, p.Dbname, p.Sslmode)
}

func getPsqlInfo(filename string) (*psqlInfo, *DatabaseError) {
	jsonFile, err := os.Open("psqlInfo.json")
	if err != nil {
		return nil, newOpenError(err)
	}
	data, err := io.ReadAll(jsonFile)
	if err != nil {
		return nil, newConfigErrror(err)
	}
	var psqlInfo psqlInfo
	err = json.Unmarshal([]byte(data), &psqlInfo)
	if err != nil {
		return nil, newConfigErrror(err)
	}
	return &psqlInfo, nil
}

func Automigrate(db *gorm.DB) *DatabaseError {
	if err := db.AutoMigrate(&schema.User{}); err != nil {
		return newMigrateError(fmt.Errorf("schema user, %w", err))
	}
	if err := db.AutoMigrate(&schema.Team{}); err != nil {
		return newMigrateError(fmt.Errorf("schema team, %w", err))
	}
	if err := db.AutoMigrate(&schema.Result{}); err != nil {
		return newMigrateError(fmt.Errorf("schema result, %w", err))
	}
	if err := db.AutoMigrate(&schema.Game{}); err != nil {
		return newMigrateError(fmt.Errorf("schema game, %w", err))
	}
	if err := db.AutoMigrate(&schema.Word{}); err != nil {
		return newMigrateError(fmt.Errorf("schema word, %w", err))
	}
	if err := db.AutoMigrate(&schema.UserDictionary{}); err != nil {
		return newMigrateError(fmt.Errorf("schema user dictionary, %w", err))
	}
	if err := db.AutoMigrate(&schema.GameWord{}); err != nil {
		return newMigrateError(fmt.Errorf("schema game word, %w", err))
	}
	if err := db.AutoMigrate(&schema.PlayerGame{}); err != nil {
		return newMigrateError(fmt.Errorf("schema player game, %w", err))
	}
	return nil
}

func AddTestUsers(db *gorm.DB) []uint {
	p1, _ := bcrypt.GenerateFromPassword([]byte("1"), bcrypt.DefaultCost)
	p2, _ := bcrypt.GenerateFromPassword([]byte("2"), bcrypt.DefaultCost)
	p3, _ := bcrypt.GenerateFromPassword([]byte("3"), bcrypt.DefaultCost)
	p4, _ := bcrypt.GenerateFromPassword([]byte("4"), bcrypt.DefaultCost)
	p5, _ := bcrypt.GenerateFromPassword([]byte("5"), bcrypt.DefaultCost)
	p6, _ := bcrypt.GenerateFromPassword([]byte("6"), bcrypt.DefaultCost)

	users := []schema.User{
		{
			Email:    "1",
			Password: p1,
			Username: "one",
		},
		{
			Email:    "2",
			Password: p2,
			Username: "two",
		},
		{
			Email:    "3",
			Password: p3,
			Username: "three",
		},
		{
			Email:    "4",
			Password: p4,
			Username: "four",
		},
		{
			Email:    "5",
			Password: p5,
			Username: "five",
		},
		{
			Email:    "6",
			Password: p6,
			Username: "six",
		},
	}
	db.Create(&users)
	ids := make([]uint, len(users))
	for i, u := range users {
		ids[i] = u.ID
	}
	return ids
}

func UpdateUser(db *gorm.DB, id uint, password []byte, username string) *DatabaseError {
	return newUpdateError(
		db.Model(&schema.User{}).
			Where("id = ?", id).
			Update("password", password).
			Update("username", username).Error,
	)
}

func UpdateUserPassword(db *gorm.DB, id uint, password []byte) *DatabaseError {
	return newUpdateError(db.Model(&schema.User{}).
		Where("id = ?", id).
		Update("password", password).Error)
}

func UpdateUserUsername(db *gorm.DB, id uint, username string) *DatabaseError {
	return newUpdateError(
		db.Model(&schema.User{}).
			Where("id = ?", id).
			Update("username", username).Error)
}

func Open(filename string) (*gorm.DB, *DatabaseError) {
	psqlInfo, derr := getPsqlInfo("psqlInfo.json")
	if derr != nil {
		return nil, derr
	}

	db, err := gorm.Open(postgres.Open(psqlInfo.String()), &gorm.Config{})
	if err != nil {
		return nil, newOpenError(err)
	}
	return db, nil
}

func AddUser(db *gorm.DB, user *schema.User) (uint, *DatabaseError) {
	if _, err := GetUserByEmail(db, user.Email); err == nil {
		return 0, newConflictError(fmt.Errorf("user with that email already exists"))
	}

	if err := db.Create(user).Error; err != nil {
		return 0, newQueryError(err)
	}
	return user.ID, nil
}

func GetUserByID(db *gorm.DB, id uint) (*schema.User, *DatabaseError) {
	var user schema.User
	err := db.First(&user, id).Error
	return &user, newQueryError(err)
}

func GetUserByEmail(db *gorm.DB, email string) (*schema.User, *DatabaseError) {
	var user schema.User
	err := db.Where("email = ?", email).First(&user).Error
	return &user, newQueryError(err)
}

func RecommendWord(db *gorm.DB, n int, id uint) ([]string, *DatabaseError) {
	rows, err := db.Raw(`
		select word, count(*) from words
		left join user_dictionaries on words.id = user_dictionaries.word_id
		where (user_dictionaries.author_id <> ? or user_dictionaries.author_id is null)
		group by words.id`, id).Rows()
	if err != nil {
		return nil, newQueryError(err)
	}

	words := make([]string, 0)
	weights := make([]float64, 0)
	sum := 0
	var word string
	var count int
	for rows.Next() {
		err = rows.Scan(&word, &count)
		if err != nil {
			return nil, newQueryError(err)
		}
		words = append(words, word)
		weights = append(weights, float64(count))

		sum += count
	}

	if len(weights) != 1 {
		for i := range weights {
			weights[i] = float64(sum) - weights[i]
		}
	}

	sampler := sampleuv.NewWeighted(
		weights,
		rand.New(rand.NewSource(uint64(time.Now().UnixNano()))),
	)
	resSize := utils.Min(len(weights), n)

	result := make([]string, resSize)
	for i := 0; i < resSize; i++ {
		index, ok := sampler.Take()
		if !ok {
			return nil, newQueryError(fmt.Errorf("fail to pick a random index"))
		}
		fmt.Printf(words[index])
		result[i] = words[index]
	}
	return result, nil
}

func AddWords(db *gorm.DB, id uint, words []string) *DatabaseError {
	return newQueryError(db.Transaction(func(tx *gorm.DB) error {
		schemaWords := make([]schema.Word, len(words))
		for i, w := range words {
			schemaWords[i] = schema.Word{
				Word: w,
			}
		}

		return nil
	}))
}

func AddGame(db *gorm.DB, game *game.Game) *DatabaseError {
	return newQueryError(db.Transaction(func(tx *gorm.DB) error {
		numTeams := int(float64(game.NumPlayers) / 2)
		schemaResults := make([]schema.Result, 0, numTeams)
		for _, r := range game.Process.Result {
			schemaTeam := schema.Team{
				FirstID:  r.FirstID,
				SecondID: r.SecondID,
			}
			if err := tx.Where(
				"first_id = ? AND second_id = ?",
				schemaTeam.FirstID,
				schemaTeam.SecondID).FirstOrCreate(&schemaTeam).Error; err != nil {
				return err
			}

			schemaResult := schema.Result{TeamID: schemaTeam.ID, Score: r.Score}

			schemaResults = append(schemaResults, schemaResult)
		}

		schemaGame := &schema.Game{
			UserID:     game.Host,
			NumPlayers: game.NumPlayers,
			Timer:      game.Timer,
			NumWords:   game.NumWords,
			Result:     schemaResults,
		}

		if err := tx.Create(schemaGame).Error; err != nil {
			return err
		}
		for userID := range game.Players.IDs {
			if err := tx.Create(&schema.PlayerGame{
				UserID: userID,
				GameID: schemaGame.ID,
			}).Error; err != nil {
				return err
			}
		}

		gameWords := make([]schema.GameWord, 0, len(game.Words.All))
		for userID, words := range game.Words.ByUser {
			for word := range words {
				schemaWord := schema.Word{Word: word}
				if err := tx.Where("word = ?", word).
					FirstOrCreate(&schemaWord).Error; err != nil {
					return err
				}
				userDictionary := schema.UserDictionary{
					AuthorID: userID,
					WordID:   schemaWord.ID,
				}
				if err := tx.Where("author_id = ? AND word_id = ?", userID, schemaWord.ID).
					FirstOrCreate(&userDictionary).Error; err != nil {
					return err
				}

				gameWords = append(gameWords, schema.GameWord{
					PlayerWordID: userDictionary.ID,
					GuessedByID:  game.Process.GuessedWords[word],
					GameID:       schemaGame.ID,
				})

			}
		}

		if err := tx.Create(gameWords).Error; err != nil {
			return err
		}

		return nil
	}))
}

func GetUserStatistics(db *gorm.DB, id uint) (containers.Statistics, *DatabaseError) {
	type Result struct {
		FirstID  uint
		SecondID uint
		Score    int
		ID       uint
	}

	words := make([]containers.Word, 0)
	var numGames int64
	var numWins int64
	var numTies int64
	var res Result
	err := db.Transaction(func(tx *gorm.DB) error {
		rows, err := tx.Model(&schema.UserDictionary{}).
			Limit(5).
			Select("words.word, count(words.word) as count").
			Joins("left join words on user_dictionaries.word_id = words.id").
			Where("author_id = ?", id).Group("words.word").
			Order("count(words.word) desc").Rows()
		if err != nil {
			return err
		}

		var word string
		var count int
		for rows.Next() {
			err = rows.Scan(&word, &count)
			if err != nil {
				return err
			}
			words = append(words, containers.Word{Word: word, Count: count})
		}

		if err := tx.Model(&schema.PlayerGame{}).
			Select("game_id").Where("user_id = ?", id).
			Count(&numGames).Error; err != nil {
			return err
		}

		rows, err = tx.Raw(`
			select teams.first_id, teams.second_id, results.score, games.id
			from game_results
			left join games on game_results.game_id = games.id
			left join results on results.id = game_results.result_id
			left join teams on teams.id = results.team_id
			where results.score = (
				select max(results2.score) from game_results as game_results2
				left join results as results2 on game_results2.result_id = results2.id
				where game_results2.game_id = games.id);`).Rows()
		if err != nil {
			return err
		}

		results := make(map[uint][]containers.Result)
		for rows.Next() {
			if err := tx.ScanRows(rows, &res); err != nil {
				return err
			}
			results[res.ID] = append(
				results[res.ID],
				containers.Result{
					FirstID:  res.FirstID,
					SecondID: res.SecondID, Score: res.Score,
				})
		}

		for _, res := range results {
			if containers.Contains(res, id) {
				if len(res) == 1 {
					numWins += 1
				} else {
					numTies += 1
				}
			}
		}

		return nil
	})
	if err != nil {
		return containers.Statistics{}, newQueryError(err)
	}
	return containers.Statistics{
		GamesPlayed:  numGames,
		NumberOfWins: numWins,
		NumberOfTies: numTies,
		TopWords:     words,
	}, nil
}
