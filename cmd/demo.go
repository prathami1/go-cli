package cmd

import (
	"github.com/prathami1/go-cli/internal/providers"
	"github.com/spf13/cobra"
)

// demoCmd represents the demo command for testing progress bars
var demoCmd = &cobra.Command{
	Use:    "demo",
	Short:  "Demo the progress bars and installation indicators",
	Long:   `Demonstrates the beautiful progress bars and spinners used during CLI installations.`,
	Hidden: true, // Hide from help since this is for development/testing
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 0 && args[0] == "download" {
			providers.DemoDownloadProgress()
		} else {
			providers.DemoProgressBars()
		}
	},
}

func init() {
	rootCmd.AddCommand(demoCmd)
}
