package main

import (
	"diplom/pkg/models"

	"github.com/labstack/gommon/log"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	logger := log.New("migrate")

	db, err := gorm.Open(sqlite.Open("db/gorm.db"), &gorm.Config{})
	must(logger, err)

	err = db.AutoMigrate(models.User{}, models.Session{})
	must(logger, err)
}

func must(logger *log.Logger, err error) {
	if err != nil {
		logger.Fatal(err)
	}
}
