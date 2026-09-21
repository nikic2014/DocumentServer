package config

import (
	"log"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	JWTSecret   string `env:"JWT_SECRET" envDefault:"dev-secret-change-me"`
	DatabaseURL string `env:"DATABASE_URL,required"`
	RedisAddr   string `env:"REDIS_ADDR,required"`
	UploadDir   string `env:"UPLOAD_DIR" envDefault:"./data/uploads"`
}

var Cfg Config

func init() {
	if err := env.Parse(&Cfg); err != nil {
		log.Fatalf("При парсинге конфига возникла ошибка: %v", err)
	}
}
