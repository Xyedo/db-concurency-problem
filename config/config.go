package config

import (
	"log"
	"os"
	"strconv"
	"sync"

	"github.com/joho/godotenv"
)

var (
	cfg  config
	once sync.Once
)

type config struct {
	Postgre struct {
		Host     string
		Port     uint64
		User     string
		Password string
		Database string
	}
}

func load(filenames ...string) {
	err := godotenv.Load(filenames...)
	if err != nil {
		log.Fatalln(err)
	}

	// Postgre
	dbPort, err := strconv.ParseUint(os.Getenv("PG_PORT"), 10, 64)
	if err != nil {
		log.Fatalln(err)
	}
	cfg.Postgre.Host = os.Getenv("PG_HOST")
	cfg.Postgre.Port = dbPort
	cfg.Postgre.User = os.Getenv("PG_USER")
	cfg.Postgre.Password = os.Getenv("PG_PASSWORD")
	cfg.Postgre.Database = os.Getenv("PG_DB")

}

func Get(filenames ...string) config {
	if cfg == (config{}) {
		once.Do(func() { load(filenames...) })
	}

	return cfg
}
