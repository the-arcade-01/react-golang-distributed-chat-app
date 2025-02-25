package main

import (
	"server/internal"

	"github.com/joho/godotenv"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		internal.Log.Error("error loading env file", "error", err)
		panic(err)
	}
	_, err = internal.ParseEnvs()
	if err != nil {
		panic(err)
	}
}

func main() {
	internal.Run()
}
