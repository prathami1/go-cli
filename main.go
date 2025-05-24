package main

import (
	"fmt"
	"os"

	"github.com/prathami1/go-cli/cmd"
)

// Build information set by ldflags
var (
	version   = "dev"
	buildTime = "unknown"
	buildUser = "unknown"
	gitCommit = "unknown"
)

func main() {
	// Check for version flag
	if len(os.Args) > 1 && (os.Args[1] == "--version" || os.Args[1] == "-v") {
		fmt.Printf("CloudDeploy CLI\n")
		fmt.Printf("Version: %s\n", version)
		fmt.Printf("Build Time: %s\n", buildTime)
		fmt.Printf("Build User: %s\n", buildUser)
		fmt.Printf("Git Commit: %s\n", gitCommit)
		return
	}

	cmd.Execute()
}
