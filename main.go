package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/bitterfly/go-chaos/hatgame/schema"
	"github.com/bitterfly/go-chaos/hatgame/server"
	_ "github.com/lib/pq"
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

func automigrate(db *gorm.DB) error {
	err := db.AutoMigrate(&schema.Users{})
	return err
}

func addTestUsers(db *gorm.DB) []uint {
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

func updateUserPassword(db *gorm.DB, id uint, password int) error {
	fmt.Printf("%d\n", id)
	return db.Model(&schema.Users{}).Where("id = ?", id).Update("password", password).Error
}

func openDB(filename string) (*gorm.DB, error) {
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

func main() {
	db, err := openDB("psqlInfo.json")
	if err != nil {
		panic(err)
	}
	log.Printf("Connected to database.")

	err = automigrate(db)
	if err != nil {
		panic(err)
	}
	log.Printf("Migrated the database.")
	// user, err := parseUser(`{"email": "foo@gmail.com", "password": 1234, "username": "dodo"}`)
	// if err != nil {
	// 	panic(err)
	// }
	// id, err := addUser(db, user)
	// if err != nil {
	// 	panic(err)
	// }

	// err = updateUserPassword(db, id, 6969)
	// if err != nil {
	// 	panic(err)
	// }
	server := server.New(db, "localhost", "8080")
	err = server.Connect()
	if err != nil {
		panic(err)
	}
}
