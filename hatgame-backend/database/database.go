package database

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/bitterfly/go-chaos/hatgame/schema"
	"github.com/bitterfly/go-chaos/hatgame/server/containers"
	"golang.org/x/crypto/bcrypt"
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

func (p psqlInfo) String() string {
	return fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=%s",
		p.Host, p.Port, p.User, p.Password, p.Dbname, p.Sslmode)
}

func getPsqlInfo(filename string) (*psqlInfo, error) {
	jsonFile, err := os.Open("psqlInfo.json")
	if err != nil {
		return nil, err
	}
	data, err := io.ReadAll(jsonFile)
	if err != nil {
		return nil, err
	}
	var psqlInfo psqlInfo
	err = json.Unmarshal([]byte(data), &psqlInfo)
	if err != nil {
		return nil, err
	}
	return &psqlInfo, nil
}

func Automigrate(db *gorm.DB) error {
	if err := db.AutoMigrate(&schema.User{}); err != nil {
		return err
	}
	if err := db.AutoMigrate(&schema.Team{}); err != nil {
		return err
	}
	if err := db.AutoMigrate(&schema.Result{}); err != nil {
		return err
	}
	if err := db.AutoMigrate(&schema.Game{}); err != nil {
		return err
	}
	if err := db.AutoMigrate(&schema.Word{}); err != nil {
		return err
	}
	if err := db.AutoMigrate(&schema.PlayerWord{}); err != nil {
		return err
	}
	if err := db.AutoMigrate(&schema.PlayerGame{}); err != nil {
		return err
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

func UpdateUser(db *gorm.DB, id uint, password []byte, username string) error {
	return db.Model(&schema.User{}).Where("id = ?", id).Update("password", password).Update("username", username).Error
}

func UpdateUserPassword(db *gorm.DB, id uint, password []byte) error {
	return db.Model(&schema.User{}).Where("id = ?", id).Update("password", password).Error
}

func UpdateUserUsername(db *gorm.DB, id uint, username string) error {
	return db.Model(&schema.User{}).Where("id = ?", id).Update("username", username).Error
}

func Open(filename string) (*gorm.DB, error) {
	psqlInfo, err := getPsqlInfo("psqlInfo.json")
	if err != nil {
		return nil, err
	}
	db, err := gorm.Open(postgres.Open(psqlInfo.String()), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return db, nil
}

func AddUser(db *gorm.DB, user *schema.User) (uint, error) {
	if err := db.Create(user).Error; err != nil {
		return 0, err
	}
	return user.ID, nil
}

func GetUserByID(db *gorm.DB, id uint) (*schema.User, error) {
	var user schema.User
	err := db.First(&user, id).Error
	return &user, err
}

func GetUserByEmail(db *gorm.DB, email string) (*schema.User, error) {
	var user schema.User
	err := db.Where("email = ?", email).First(&user).Error
	return &user, err
}

func AddGame(db *gorm.DB, game *containers.Game) error {
	return db.Transaction(func(tx *gorm.DB) error {
		playerWords := make([]schema.PlayerWord, 0, len(game.Players.Words))
		for userId, words := range game.Players.WordsByUser {
			for word := range words {
				schemaWord := schema.Word{Word: word}
				if err := tx.Where("word = ?", word).FirstOrCreate(&schemaWord).Error; err != nil {
					return err
				}

				playerWords = append(
					playerWords, schema.PlayerWord{
						AuthorID:    userId,
						GuessedByID: game.Process.GuessedWords[word],
						WordID:      schemaWord.ID})
			}
		}

		if err := tx.Create(playerWords).Error; err != nil {
			return err
		}

		numTeams := int(float64(game.NumPlayers) / 2)
		schemaResults := make([]schema.Result, 0, numTeams)
		for _, r := range game.Process.Result {
			schemaTeam := schema.Team{
				FirstID:  r.FirstID,
				SecondID: r.SecondID,
			}
			if err := tx.Where("first_id = ? AND second_id = ?", schemaTeam.FirstID, schemaTeam.SecondID).FirstOrCreate(&schemaTeam).Error; err != nil {
				return err
			}

			schemaResult := schema.Result{TeamID: schemaTeam.ID, Score: r.Score}

			schemaResults = append(schemaResults, schemaResult)
		}

		schemaGame := &schema.Game{
			UserID:      game.Host,
			NumPlayers:  game.NumPlayers,
			Timer:       game.Timer,
			NumWords:    game.NumWords,
			PlayerWords: playerWords,
			Result:      schemaResults,
		}

		if err := tx.Create(schemaGame).Error; err != nil {
			return err
		}
		for userID := range game.Players.Ws {
			if err := tx.Create(&schema.PlayerGame{
				UserID: userID,
				GameID: schemaGame.ID,
			}).Error; err != nil {
				return err
			}
		}

		return nil
	})

}

type Result struct {
	FirstID  uint
	SecondID uint
	Score    int
	ID       uint
}

func GetUserStatistics(db *gorm.DB, id uint) (containers.Statistics, error) {
	words := make([]containers.Word, 0)
	var numGames int64
	var numWins int64
	var numTies int64
	var res Result
	err := db.Transaction(func(tx *gorm.DB) error {
		rows, err := tx.Model(&schema.PlayerWord{}).Limit(5).Select("words.word, count(words.word) as count").Joins("left join words on player_words.word_id = words.id").Where("author_id = ?", id).Group("words.word").Order("count(words.word) desc").Rows()
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

		if err := tx.Model(&schema.PlayerGame{}).Select("game_id").Where("user_id = ?", id).Count(&numGames).Error; err != nil {
			return err
		}

		rows, err = tx.Raw("select teams.first_id, teams.second_id, results.score, games.id from game_results left join games on game_results.game_id = games.id left join results on results.id = game_results.result_id left join teams on teams.id = results.team_id where results.score = (select max(results2.score) from game_results as game_results2 left join results as results2 on game_results2.result_id = results2.id where game_results2.game_id = games.id);").Rows()
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
					SecondID: res.SecondID, Score: res.Score})
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
		return containers.Statistics{}, err
	}
	return containers.Statistics{
		GamesPlayed:  numGames,
		NumberOfWins: numWins,
		NumberOfTies: numTies,
		TopWords:     words,
	}, nil
}
