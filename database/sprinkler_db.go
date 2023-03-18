// Copyright (c) 2022, Salesforce, Inc.
// All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause
// For full license text, see the LICENSE file in the repo root or https://opensource.org/licenses/BSD-3-Clause

package database

import (
	"fmt"
	"sync"

	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Dsn struct {
	Host     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

func (dsn *Dsn) stringify() string {
	return fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s sslmode=%s",
		dsn.Host,
		dsn.User,
		dsn.Password,
		dsn.DBName,
		dsn.SSLMode,
	)
}

// singleton
var once sync.Once
var sprinklerDB *gorm.DB

func GetInstance() *gorm.DB {
	once.Do(func() {
		dsn := &Dsn{
			Host:     viper.GetString("db.host"),
			User:     viper.GetString("db.user"),
			Password: viper.GetString("db.password"),
			DBName:   viper.GetString("db.dbname"),
			SSLMode:  viper.GetString("db.sslmode"),
		}

		db, err := gorm.Open(postgres.Open(dsn.stringify()), &gorm.Config{})
		if err != nil {
			panic("failed to connect database")
		}
		sprinklerDB = db
	})
	return sprinklerDB
}
