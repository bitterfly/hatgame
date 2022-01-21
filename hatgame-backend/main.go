package main

import (
	"log"

	"github.com/bitterfly/go-chaos/hatgame/database"
	"github.com/bitterfly/go-chaos/hatgame/server"
	_ "github.com/lib/pq"
)

func main() {
	db, err := database.Open("psqlInfo.json")
	if err != nil {
		panic(err)
	}
	log.Printf("Connected to database.")

	err = database.Automigrate(db)
	if err != nil {
		panic(err)
	}
	log.Printf("Migrated the database.")

	server := server.New(db)
	err = server.Connect("localhost:8080")
	if err != nil {
		panic(err)
	}
}
