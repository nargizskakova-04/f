package config

import (
	"encoding/json"
	"os"
	"time"
)

func GetConfig(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var cfg Config
	if err = json.NewDecoder(f).Decode(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

type Config struct {
	App        App        `json:"app"`
	Repository Repository `json:"repository"`
}

type App struct {
	Port int           `json:"port"`
	RTO  time.Duration `json:"rto"`
	WTO  time.Duration `json:"wto"`
}

type Repository struct {
	DBHost      string `json:"db_host"`
	DBSrv       string `json:"db_srv"`
	DBPort      int    `json:"db_port"`
	DBUsername  string `json:"db_username"`
	DBPassword  string `json:"db_password"`
	DBName      string `json:"db_name"`
	DBSSLMode   string `json:"db_ssl_mode"`
	MaxConn     int    `json:"max_conn"`
	MaxIdleConn int    `json:"max_idle_conn"`
}
