package config

import (
	"fmt"
	"os"
)

type AppConfig struct {
	DBHOST string
	DBPORT string
	DBUSER string
	DBPASS string
	DBNAME string
}

func LoadConfig() (*AppConfig, error) {

	cfg := &AppConfig{
		DBHOST: os.Getenv("DB_HOST"),
		DBPORT: os.Getenv("DB_PORT"),
		DBUSER: os.Getenv("DB_USER"),
		DBPASS: os.Getenv("DB_PASSWORD"),
		DBNAME: os.Getenv("DB_NAME"),
	}

	if cfg.DBHOST == "" || cfg.DBUSER == "" || cfg.DBNAME == "" {
		return nil, fmt.Errorf("missing required database configuration: DB_HOST, DB_USER, and DB_NAME must be set")
	}

	return cfg, nil
}

func (c *AppConfig) GetDSN() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		c.DBHOST, c.DBPORT, c.DBUSER, c.DBPASS, c.DBNAME)
}
