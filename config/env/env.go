package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)


func Load() {
	err := godotenv.Load()
	if err != nil {
		// log the error if .env file is not found or cannot be loaded
		fmt.Println("err loading .env file")
	}
}

func getKey(key string) any {
	value, ok := os.LookupEnv(key)

	if !ok {
		return ""
	}
	return value
}

func GetString(key string, fallback string) string {
	value := getKey(key)

	if value == "" {
		return fallback
	}
	return value.(string)
}

func GetInt(key string, fallback int) int {

	value := getKey(key)

	if value == "" {
		return fallback
	}

	intValue, err := strconv.Atoi(value.(string))
	if err != nil {
		fmt.Printf("Error converting %s to int: %v\n", key, err)
		return fallback
	}

	return intValue
}

func GetBool(key string, fallback bool) bool {

	value := getKey(key)

	if value == "" {
		return fallback
	}

	boolValue, err := strconv.ParseBool(value.(string))
	if err != nil {
		fmt.Printf("Error converting %s to bool: %v\n", key, err)
		return fallback
	}
	return boolValue
}

func GetFloat(key string, fallback float64) float64 {
	value := getKey(key)

	if value == "" {
		return fallback
	}

	floatValue, err := strconv.ParseFloat(value.(string), 64)
	if err != nil {
		fmt.Printf("Error converting %s to float: %v\n", key, err)
		return fallback
	}
	return floatValue
}
