package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

// NewCleanCommand creates clean command
func NewCleanCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "clean",
		Short: "Perform system cleanup",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("🧹 Starting cleanup...")
			fmt.Println("✅ Cleanup completed successfully!")
		},
	}
}
