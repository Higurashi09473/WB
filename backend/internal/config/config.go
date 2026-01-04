// Package config contains configuration structures and loading logic for the WB backend service.
// It supports reading configuration from YAML files and environment variables using cleanenv.
package config

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

// Config represents the root application configuration.
type Config struct {
	Env            string `yaml:"env" env-default:"local"`
	StoragePath    string `yaml:"storage_path" env-required:"true"`
	MigrationsPath string `yaml:"migrations_path"`
	HTTPServer     `yaml:"http_server"`
	Postgresql     `yaml:"postgresql"`
	Redis          `yaml:"redis"`
	Kafka          `yaml:"kafka"`
}

// HTTPServer holds HTTP server configuration.
type HTTPServer struct {
	Address     string        `yaml:"address" env-default:"localhost:8080"`
	Timeout     time.Duration `yaml:"timeout" env-default:"10s"`
	IdleTimeout time.Duration `yaml:"env" env-default:"60s"`
}

// Postgresql contains PostgreSQL connection settings.
type Postgresql struct {
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DBname   string `yaml:"dbname"`
	SSLmode  string `yaml:"sslmode" env-default:"disable"`
	Host     string `yaml:"host" env-default:"localhost"`
	Port     int    `yaml:"port" env-default:"5432"`
}

// Redis contains Redis connection settings.
type Redis struct {
	Host     string `yaml:"host" env:"REDIS_HOST" env-default:"localhost"`
	Port     string `yaml:"port" env:"REDIS_PORT" env-default:"6379"`
	Password string `yaml:"password" env:"REDIS_PASSWORD"`
	DB       int    `yaml:"db" env:"REDIS_DB" env-default:"0"`
}

// Kafka contains Kafka broker and topic configuration.
type Kafka struct {
	Brokers       []string `yaml:"brokers"`
	ConsumerGroup string   `yaml:"consumer_group"`
	Topic         string   `yaml:"topic"`
	DLQTopic      string   `yaml:"dlq_topic"`
}

// MustLoad loads configuration from YAML file and environment variables.
// It panics if the config file is missing or cannot be read.
func MustLoad() *Config {
	configPath := "./configs/local.yaml" 
	if configPath == "" {
		log.Fatal("CONFIG_PATH is not set")
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("config file does not exist: %s", configPath)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("cannot read config: %s", err)
	}

	return &cfg
}

// DSN returns PostgreSQL connection string in the format required by pgx/driver.
func (p Postgresql) DSN() string {
	return fmt.Sprintf("user=%s password=%s dbname=%s sslmode=%s host=%s port=%d",
        p.User, p.Password, p.DBname, p.SSLmode, p.Host, p.Port)
}