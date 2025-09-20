package main

import (
	"go-lambda-api/internal/app"
)

func main() {
	// Set environment variable to trigger local server mode
	// os.Setenv("GO_RUN_MAIN", "1") // Removed to allow Lambda path testing

	// Call the AppMain function from the app package
	app.Main()
}
