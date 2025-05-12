package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port          string
	AddTime       int
	Subtime       int
	MultiplicTime int
	Divtime       int
}

func getEnv(key string, defaultValue int) int {
	value, _ := strconv.Atoi(os.Getenv(key))
	if value == 0 {
		return defaultValue
	}
	return value
}

func ConfigFromEnv() *Config {
	return &Config{
		Port:          strconv.Itoa(getEnv("PORT", 8081)),
		AddTime:       getEnv("TIME_ADDITION_MS", 10),
		Subtime:       getEnv("TIME_SUBTRACTION_MS", 10),
		MultiplicTime: getEnv("TIME_MULTIPLICATIONS_MS", 10),
		Divtime:       getEnv("TIME_DIVISIONS_MS", 10),
	}
}
