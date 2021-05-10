package main

import "os"

func folderExists(folderPath string) bool {
	_, err := os.Stat(folderPath)

	return !os.IsNotExist(err)
}

func createSettings() {
	// TO DO:
	// Create JSON file with default settings

	type DefaultSettings struct {
		AuthenticationCookie string
	}
}

func loadSettings() {
	// Check if JSON file exists
	// If so, load into struct
	// Else, create settings and return
}
