package utils

import (
	"os"
)

func GetToken() string {
	token := os.Getenv("GITHUB_TOKEN")
	if token != "" {
		return token
	}
	return os.Getenv("GH_TOKEN")
}
