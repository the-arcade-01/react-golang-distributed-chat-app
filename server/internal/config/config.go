package config

import (
	"database/sql"
	"log"
	"os"
	"sync"

	_ "github.com/go-sql-driver/mysql"
	"github.com/redis/go-redis/v9"
)

var once sync.Once
var config *AppConfig

type AppConfig struct {
	Db    *sql.DB
	Cache *redis.Client
}

func NewAppConfig() *AppConfig {
	once.Do(func() {
		config = &AppConfig{
			Db:    newDbClient(),
			Cache: newCacheClient(),
		}
	})
	return config
}

func newDbClient() *sql.DB {
	db, err := sql.Open(os.Getenv("DB_DRIVER"), os.Getenv("DB_URL"))
	if err != nil {
		log.Fatalf("[newDbClient] error on establishing db conn, err: %v\n", err)
	}
	if err := db.Ping(); err != nil {
		log.Fatalf("[newDbClient] error on ping db, err: %v\n", err)
	}
	log.Println("[newDbClient] db conn established")
	return db
}

func newCacheClient() *redis.Client {
	cache := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDR"),
		Password: os.Getenv("REDIS_PWD"),
		DB:       0,
	})
	log.Println("[newCacheClient] cache conn established")
	return cache
}
