package internal

import (
	"fmt"
	"os"
	"strconv"
	"sync"
)

var envOnce sync.Once
var Envs *AppEnvs

type AppEnvs struct {
	ENV                   string
	REDIS_ADDR            string
	REDIS_PWD             string
	REDIS_DB              int
	CHAT_CHANNEL          string
	WEB_URL               string
	WS_TYPE               string
	STREAM_KEY            string
	STREAM_CONSUMER_GROUP string
	MAX_CHAT_LEN          int
}

func ParseEnvs() (*AppEnvs, error) {
	var appErr error
	envOnce.Do(func() {
		Envs = &AppEnvs{
			ENV:                   os.Getenv("ENV"),
			REDIS_ADDR:            os.Getenv("REDIS_ADDR"),
			REDIS_PWD:             os.Getenv("REDIS_PWD"),
			CHAT_CHANNEL:          os.Getenv("CHAT_CHANNEL"),
			WEB_URL:               os.Getenv("WEB_URL"),
			WS_TYPE:               os.Getenv("WS_TYPE"),
			STREAM_KEY:            os.Getenv("STREAM_KEY"),
			STREAM_CONSUMER_GROUP: os.Getenv("STREAM_CONSUMER_GROUP"),
		}

		if Envs.ENV == "" || Envs.REDIS_ADDR == "" || Envs.REDIS_PWD == "" || Envs.CHAT_CHANNEL == "" || Envs.WEB_URL == "" || Envs.STREAM_KEY == "" || Envs.STREAM_CONSUMER_GROUP == "" {
			appErr = fmt.Errorf("invalid env variables, please check .env file")
			return
		}

		db, err := stringToInt(os.Getenv("REDIS_DB"))
		if err != nil {
			appErr = err
			return
		}
		Envs.REDIS_DB = db

		length, err := stringToInt(os.Getenv("MAX_CHAT_LEN"))
		if err != nil {
			appErr = err
			return
		}
		Envs.MAX_CHAT_LEN = length
	})

	if appErr != nil {
		return nil, appErr
	}

	Log.Info("env variables loaded successfully")
	Log.Info("active environment", "env", Envs.ENV)

	return Envs, nil
}

func stringToInt(s string) (int, error) {
	val, err := strconv.Atoi(s)
	if err != nil {
		Log.Error("error on converting string to int", "param", s, "error", err)
		return -1, err
	}
	return val, nil
}
