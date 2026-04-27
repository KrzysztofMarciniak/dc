package models

import (
	"bufio"
	"log"
	"os"
	"strings"
)

type Config struct {
	AdminUsername string
	AdminPassword string
}

func LoadConfig() Config {
	loadDotEnv(".env")

	cfg := Config{
		AdminUsername: os.Getenv("ADMIN_USERNAME"),
		AdminPassword: os.Getenv("ADMIN_PASSWORD"),
	}

	if cfg.AdminUsername == "" || cfg.AdminPassword == "" {
		log.Fatal("missing required env vars: ADMIN_USERNAME or ADMIN_PASSWORD")
	}

	return cfg
}
func loadDotEnv(path string) {
	file, err := os.Open(path)
	if err != nil {
		log.Fatal("cannot open .env file:", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// ignore comments / empty lines
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		os.Setenv(key, value)
	}

	if err := scanner.Err(); err != nil {
		log.Fatal("error reading .env:", err)
	}
}
