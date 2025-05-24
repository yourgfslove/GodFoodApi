package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"log"
	"os"
	"time"
)

type Config struct {
	Env        string `yaml:"env" env:"ENV" env-default:"local"`
	StorageURL string `yaml:"storage_url" env:"STORAGE_URL" env-required:"true"`
	SecretJWT  string `yaml:"secret_jwt" env:"SECRET_JWT"`
	HTTPServer `yaml:"http_server" env:"HTTP_SERVER" env-required:"true"`
}
type HTTPServer struct {
	Address     string        `yaml:"address" env-default:"localhost:8081"`
	Timeout     time.Duration `yaml:"timeout" env-default:"5s"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env-default:"60s"`
}

func MustLoadConfig() *Config {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		log.Fatal("CONFIG_PATH environment variable not set")
	}
	_, err := os.Stat(configPath)
	if err != nil {
		log.Fatalf("%s does not exist: %v", configPath, err)
	}
	var config Config
	if err = cleanenv.ReadConfig(configPath, &config); err != nil {
		log.Fatalf("Error reading config: %v", err)
	}
	return &config
}
