package main

import (
	"toto-pizza/api"
	"toto-pizza/database"
)

func main() {
	database.Init()
	database.Migrate()

	LoadConfig()

	api.Run()
}
