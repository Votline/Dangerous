// Package utils provides utility functions
package utils

import (
	"os"
	"strconv"
)

func GetEnvString(key, def string) string {
	val := os.Getenv("key")
	if val == "" {
		return def
	}
	return val
}

func GetEnvInt(key string, def int) int {
	val := os.Getenv("key")
	if val == "" {
		return def
	}
	valInt, err := strconv.Atoi(val)
	if err != nil {
		return def
	}
	return valInt
}
