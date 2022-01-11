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
		schemaWords := make([]*schema.Word, 0, len(game.Players.Words))
		schemaWordsMap := make(map[string]*schema.Word)
		for word := range game.Players.Words {
			schemaWords = append(schemaWords, &schema.Word{Word: word})
			schemaWordsMap[word] = schemaWords[len(schemaWords)-1]
		}

		if err := tx.Create(schemaWords).Error; err != nil {
			return err
		}

		playerWords := make([]schema.PlayerWord, 0, len(game.Players.Words))
		for userId, words := range game.Players.WordsByUser {
			for word := range words {
				playerWords = append(
					playerWords, schema.PlayerWord{UserID: userId, WordID: schemaWordsMap[word].ID})
			}
		}

		if err := tx.Create(playerWords).Error; err != nil {
			return err
		}

		schemaGame := &schema.Game{
			UserID:      game.Host,
			NumPlayers:  game.NumPlayers,
			Timer:       game.Timer,
			NumWords:    game.NumWords,
			PlayerWords: playerWords,
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

// db.Model(&data).Association("Entities").Append([]*Entity{&Entity{Name: "mynewentity"}})
