package config

import (
	"flag"

	"github.com/caarlos0/env/v6"
)

type accrualServerConfig struct {
	Host        string `env:"RUN_ADDRESS"`
	DatabaseURI string `env:"DATABASE_URI"`
}

func GetAccrualServerConfig() (host, dbURI *string) {

	var cfg accrualServerConfig

	host, dbURI = getServerFlags()

	_ = env.Parse(&cfg)

	flag.Parse()

	if cfg.Host != "" {
		host = &cfg.Host
	}

	if cfg.DatabaseURI != "" {
		dbURI = &cfg.DatabaseURI
	}

	return
}

// host=localhost user=postgres password=postgres sslmode=disable dbname=accrual
func getServerFlags() (host, dbURI *string) {
	host = flag.String("a", "localhost:8085", "host address")
	dbURI = flag.String("d", "host=localhost user=postgres password=postgres sslmode=disable dbname=accrual", "database address")
	return
}
