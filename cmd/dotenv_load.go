package main

import (
	"os"
	"strings"

	"github.com/joho/godotenv"
)

func dotEnvPathExists(path string) bool {
	st, err := os.Stat(path)
	return err == nil && !st.IsDir()
}

// loadDotEnvFiles loads .env from the current working directory.
// Set KUAKE_LOAD_DOTENV=0 to disable.
func loadDotEnvFiles() {
	if strings.TrimSpace(os.Getenv("KUAKE_LOAD_DOTENV")) == "0" {
		return
	}
	if dotEnvPathExists(".env") {
		_ = godotenv.Load(".env")
	}
}
