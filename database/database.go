package database

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/bitterfly/go-chaos/hatgame/schema"
	"github.com/bitterfly/go-chaos/hatgame/server/containers"
	"github.com/bitterfly/go-chaos/hatgame/utils"
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
		fmt.Printf("%d\n", u.ID)
		ids[i] = u.ID
	}
	return ids
}

func UpdateUserPassword(db *gorm.DB, id uint, password []byte) error {
	fmt.Printf("%d\n", id)
	return db.Model(&schema.User{}).Where("id = ?", id).Update("password", password).Error
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
		fmt.Printf("NumPlayers: %d\n", numTeams)
		schemaTeams := make([]schema.Team, 0, numTeams)
		for i := 0; i < numTeams; i++ {
			firstID, secondID := utils.Order(
				game.Process.Teams[i],
				game.Process.Teams[(i+numTeams)%game.NumPlayers])
			schemaTeam := schema.Team{
				FirstID:  firstID,
				SecondID: secondID,
			}
			if err := tx.Where("first_id = ? AND second_id = ?", schemaTeam.FirstID, schemaTeam.SecondID).FirstOrCreate(&schemaTeam).Error; err != nil {
				return err
			}

			schemaTeams = append(schemaTeams, schemaTeam)
		}
		fmt.Printf("%d, %d", game.Process.WinningTeam.First, game.Process.WinningTeam.Second)

		var winningTeamID uint
		if err := tx.Model(&schema.Team{}).Select("id").Where("first_id = ? AND second_id = ?", game.Process.WinningTeam.First, game.Process.WinningTeam.Second).First(&winningTeamID).Error; err != nil {
			return err
		}

		schemaGame := &schema.Game{
			UserID:      game.Host,
			NumPlayers:  game.NumPlayers,
			Timer:       game.Timer,
			NumWords:    game.NumWords,
			PlayerWords: playerWords,
			Teams:       schemaTeams,
			TeamID:      winningTeamID,
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

type result struct {
	Word  string
	Count int
}

func GetUserStatistics(db *gorm.DB, id uint) (containers.Statistics, error) {
	words := make([]containers.Word, 0)
	var numGames int64
	var numWins int64
	err := db.Transaction(func(tx *gorm.DB) error {
		rows, err := tx.Model(&schema.PlayerWord{}).Limit(10).Select("words.word, count(words.word) as count").Joins("left join words on player_words.word_id = words.id").Where("author_id = ?", id).Group("words.word").Order("count(words.word) desc").Rows()
		if err != nil {
			fmt.Printf("%s", err.Error())
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

		if err := tx.Model(&schema.PlayerGame{}).Distinct("game_id").Where("user_id = ?", id).Count(&numGames).Error; err != nil {
			return err
		}

		if err := tx.Model(&schema.Game{}).Joins("left join teams on games.team_id = teams.id").Where("teams.first_id = ?", id).Or("teams.second_id = ?", id).Count(&numWins).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return containers.Statistics{}, err
	}
	return containers.Statistics{
		GamesPlayed:  numGames,
		NumberOfWins: numWins,
		TopWords:     words,
	}, nil
}
