package main

import (
	"encoding/json"
	"fmt"
	"os"
	"toto-pizza/database"

	"gorm.io/gorm"
)

// Item представляет структуру элемента меню
type Item struct {
	gorm.Model
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        string `json:"type"`
	Cost        int    `json:"cost"`
	Weight      int    `json:"weight"`
}

// Config представляет структуру JSON-файла
type Config struct {
	Menu []Item `json:"menu"`
}

func LoadConfig() {
	// Открываем и читаем JSON файл
	file, err := os.ReadFile("config.json")
	if err != nil {
		panic(fmt.Sprintf("Не удалось открыть файл: %v", err))
	}

	// Парсим JSON
	var config Config
	err = json.Unmarshal(file, &config)
	if err != nil {
		panic(fmt.Sprintf("Не удалось распарсить JSON: %v", err))
	}

	// Загружаем данные в базу
	for _, item := range config.Menu {
		database.DB.Create(&database.Item{Name: item.Name, Description: item.Description, Type: item.Type, Weight: uint(item.Weight), Cost: uint(item.Cost)})
	}

	fmt.Println("Загрузка данных завершена!")
}
