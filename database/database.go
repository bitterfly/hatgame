package database

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/bitterfly/go-chaos/hatgame/schema"
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
	err := db.AutoMigrate(&schema.Users{})
	return err
}

func AddTestUsers(db *gorm.DB) []uint {
	users := []schema.Users{
		{Email: "dodo@gmail.com", Password: 1234, Username: "dodo"},
		{Email: "foo@gmail.com", Password: 4567, Username: "foo"},
		{Email: "bar@gmail.com", Password: 8910, Username: "bar"},
	}
	db.Create(&users)
	ids := make([]uint, len(users))
	for i, u := range users {
		fmt.Printf("%d\n", u.ID)
		ids[i] = u.ID
	}
	return ids
}

func UpdateUserPassword(db *gorm.DB, id uint, password int) error {
	fmt.Printf("%d\n", id)
	return db.Model(&schema.Users{}).Where("id = ?", id).Update("password", password).Error
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

func AddUser(db *gorm.DB, user *schema.Users) (uint, error) {
	if err := db.Create(user).Error; err != nil {
		return 0, err
	}
	return user.ID, nil
}
