package database

import "gorm.io/gorm"

type Order struct {
	gorm.Model
	UserId  uint
	Cost    uint
	Status  string
	Address string
}

type OrderItem struct {
	gorm.Model
	OrderId uint
	Cost    uint
	ItemId  uint
	Count   uint
}

type User struct {
	gorm.Model
	Name         string
	PasswordHash string
	Phone        string `gorm:"uniqueIndex"`
	Balance      uint
}

type Photo struct {
	gorm.Model
	ItemId   uint
	Link     string `gorm:"uniqueIndex"`
	Position uint
}

type Item struct {
	gorm.Model
	Cost        uint
	Name        string `gorm:"uniqueIndex"`
	Description string
	Type        string
	Weight      uint
}

type Session struct {
	gorm.Model
	UserId    uint
	SessionId string
}
