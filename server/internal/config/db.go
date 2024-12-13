package config

import (
	"log"
	"os"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func newDBClient() (*gorm.DB, error) {
	dbUrl := os.Getenv("DB_URL")
	db, err := gorm.Open(mysql.Open(dbUrl), &gorm.Config{})
	if err != nil {
		log.Fatalf("[newDBClient] error connecting to db, %v\n", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("[newDBClient] error getting sql.DB from gorm.DB, %v\n", err)
	}
	if err := sqlDB.Ping(); err != nil {
		log.Fatalf("[newDBClient] error pinging db, %v\n", err)
	}
	return db, nil
}
