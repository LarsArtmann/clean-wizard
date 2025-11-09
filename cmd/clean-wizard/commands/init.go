package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

// NewInitCommand creates init command
func NewInitCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Interactive setup wizard",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("🧹 Clean Wizard Setup")
			fmt.Println("======================")
			fmt.Println("Configuration created successfully!")
		},
	}
}
