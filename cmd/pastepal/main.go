package main

import (
	"fmt"
	"os"

	"github.com/JacobRWebb/PastePal-OS/internal/core"
)

func main() {
	// Initialize the application
	configPath := core.GetConfigPath()
	app, err := core.NewApp(configPath)
	if err != nil {
		fmt.Printf("Error initializing application: %v\n", err)
		os.Exit(1)
	}

	// Create and run the GUI
	gui := NewGUI(app)
	gui.Run()
}
