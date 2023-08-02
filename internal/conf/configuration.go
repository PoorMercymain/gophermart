package conf

import (
	"flag"

	"github.com/caarlos0/env/v6"
)

type Config struct {
	ServerAddress  string `env:"RUN_ADDRESS"`
	DatabaseURI    string `env:"DATABASE_URI"`
	MongoURI       string `env:"MONGO_URI"`
	AccrualAddress string `env:"ACCRUAL_ADDRESS"`
}

func GetServerConfig() (outCfg *Config) {

	var envCfg Config

	outCfg = getServerFlags()

	_ = env.Parse(&envCfg)

	flag.Parse()

	if envCfg.ServerAddress != "" {
		outCfg.ServerAddress = envCfg.ServerAddress
	}

	if envCfg.DatabaseURI != "" {
		outCfg.DatabaseURI = envCfg.DatabaseURI
	}

	if envCfg.MongoURI != "" {
		outCfg.MongoURI = envCfg.MongoURI
	}

	if envCfg.AccrualAddress != "" {
		outCfg.AccrualAddress = envCfg.AccrualAddress
	}

	return
}

func getServerFlags() (cfg *Config) {
	cfg = &Config{}
	cfg.ServerAddress = *flag.String("a", "http://localhost:8080", "server address")
	cfg.DatabaseURI = *flag.String("d", "host=localhost dbname=gophermart-postgres user=gophermart-postgres password=gophermart-postgres port=3000 sslmode=disable", "postgres DSN")
	cfg.MongoURI = *flag.String("m", "mongodb://mongodb:27017", "Mongo URI")
	cfg.AccrualAddress = *flag.String("c", "http://localhost:8085", "accrual server address")
	return
}
