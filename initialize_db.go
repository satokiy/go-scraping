package main

import (
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type DBsettings struct {
	Host string
	Port string
	Name string
	User string
	Pass string
}

func connDB() (*gorm.DB, error) {
	s := DBsettings{
		Host: "localhost",
		Port: "4306",
		Name: "go_scraping_dev",
		User: "go-scraping-user",
		Pass: "password",
	}

	dsn := fmt.Sprintf("%s:%s@(%s:%s)/%s?charset=utf8&parseTime=true&loc=Local", s.User, s.Pass, s.Host, s.Port, s.Name)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("db connection error: %v", err)
	}
	return db, nil
}

func migrateDB(db *gorm.DB) error {
	if err := db.AutoMigrate(&ItemMaster{}, &LatestItem{}); err != nil {
		return fmt.Errorf("db migration error: %v", err)
	}

	return nil
}
