package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Product struct {
	gorm.Model
	Code  string
	Price uint
}

func main() {
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold: time.Second, // Slow SQL threshold
			LogLevel:      logger.Info, // Log level
		},
	)

	dsn := "host=db user=postgres password=sprinkler dbname=postgres sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		DryRun: true,
		Logger: newLogger,
	})
	if err != nil {
		panic("failed to connect database")
	}

	db.Migrator().CreateTable(&Product{})
	fmt.Println("Hello World!")
}
