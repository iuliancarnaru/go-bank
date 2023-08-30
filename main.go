package main

import (
	"log"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	store, err := NewPostgresStore()
	if err != nil {
		log.Fatal(err)
	}

	// fmt.Printf("%+v\n", store)
	if err := store.Init(); err != nil {
		log.Fatal(err)
	}

	server := NewAPIServer(":4000", store)
	server.Run()
}
