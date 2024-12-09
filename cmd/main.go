package main

import (
	"github.com/Kapeland/task-Astral/internal/app"
	"log"
)

func main() {
	log.Print("app started")
	err := app.Start()
	if err != nil {
		log.Fatal(err)
	}
	log.Print("app finished")
}
