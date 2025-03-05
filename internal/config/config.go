package config

import (
	"os"
	"strconv"
)

type Config struct {
	ComputingPower       int
	TimeAdditionMS       int
	TimeSubtractionMS    int
	TimeMultiplicationMS int
	TimeDivisionMS       int
}

func Load() *Config {
	return &Config{
		ComputingPower:       getEnvAsInt("COMPUTING_POWER", 2),
		TimeAdditionMS:       getEnvAsInt("TIME_ADDITION_MS", 10000),
		TimeSubtractionMS:    getEnvAsInt("TIME_SUBTRACTION_MS", 10000),
		TimeMultiplicationMS: getEnvAsInt("TIME_MULTIPLICATION_MS", 10000),
		TimeDivisionMS:       getEnvAsInt("TIME_DIVISION_MS", 10000),
	}
}

func getEnvAsInt(name string, defaultVal int) int {
	valStr := os.Getenv(name)
	if val, err := strconv.Atoi(valStr); err == nil {
		return val
	}
	return defaultVal
}
