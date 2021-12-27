package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

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

func main() {
	psqlInfo, err := getPsqlInfo("psqlInfo.json")
	if err != nil {
		panic(err)
	}
	db, err = gorm.Open(postgres.Open(psqlInfo.String()), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	fmt.Println("Successfully connected!")
}
