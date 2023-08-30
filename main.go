package main

import (
	"log"
)

func main() {
	store, err := NewPostgresStore()
	if err != nil {
		log.Fatal(err)
	}

	// fmt.Printf("%+v\n", store)

	server := NewAPIServer(":4000", store)
	server.Run()
}
