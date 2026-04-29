package main

import (
	"_discord_bot/db"
	"_discord_bot/router"
	"_discord_bot/templates"
	"log"
)

func main() {
	if err := db.Init(); err != nil {
		log.Fatal(err)
	}
	r := router.New()
	templates.Init(r)

	r.Run(":8080")
}
