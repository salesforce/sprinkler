package database

import (
	"sync"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var dsn = "host=db user=postgres password=sprinkler dbname=postgres sslmode=disable"

// singleton
var once sync.Once
var sprinklerDB *gorm.DB

func GetInstance() *gorm.DB {
	once.Do(func() {
		db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err != nil {
			panic("failed to connect database")
		}
		sprinklerDB = db
	})
	return sprinklerDB
}
