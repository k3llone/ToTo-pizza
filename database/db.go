package database

import (
	"log"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Init() {
	var err error
	DB, err = gorm.Open(sqlite.Open("data.db"), &gorm.Config{})

	if err != nil {
		log.Panicln(err)
	}
}

func Migrate() {
	DB.AutoMigrate(&User{})
	DB.AutoMigrate(&Order{})
	DB.AutoMigrate(&OrderItem{})
	DB.AutoMigrate(&Item{})
	DB.AutoMigrate(&Photo{})
	DB.AutoMigrate(&Session{})
}
