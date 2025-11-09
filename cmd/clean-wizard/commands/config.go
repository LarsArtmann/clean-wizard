package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

// NewConfigCommand creates config command
func NewConfigCommand() *cobra.Command {
	var configCmd = &cobra.Command{
		Use:   "config",
		Short: "Manage configuration",
	}

	configCmd.AddCommand(&cobra.Command{
		Use:   "show",
		Short: "Show current configuration",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Configuration file: ~/.clean-wizard.yaml")
			fmt.Println("Version: dev")
			fmt.Println("Safe mode: true")
			fmt.Println("Dry run: true")
		},
	})

	return configCmd
}
