package config

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type projectEnv struct {
	PostgresUser string
	PostgresPass string
	PostgresDB   string
}


func ParseEnv() (*projectEnv, error){
	var sErr []string


	err := godotenv.Load("./.env.example")
	if err != nil {
		log.Fatal("Ошибка загрузки .env файла")
	}

	PostgresUser := os.Getenv("PostgresUser")
	if PostgresUser == "" {
		sErr = append(sErr, "invalid env for PostgresUser")
	}

	PostgresPass := os.Getenv("PostgresPass")
	if PostgresPass == "" {
		sErr = append(sErr, "invalid env for PostgresPass")
	}


	PostgresDB := os.Getenv("PostgresDB")
	if PostgresDB == "" {
		sErr = append(sErr, "invalid env for PostgresDB")
	}
	
	if len(sErr) > 0 {
		return nil, fmt.Errorf("%s", strings.Join(sErr, ", "))
	}

	return &projectEnv{
		PostgresPass: PostgresPass,
		PostgresUser: PostgresUser,
		PostgresDB: PostgresDB,
	}, nil
}