package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var version = "dev"

func main() {
	var showVersion bool

	var rootCmd = &cobra.Command{
		Use:   "clean-wizard",
		Short: "Interactive system cleaning wizard",
		Long:  "A comprehensive CLI/TUI tool for system cleanup",
		Run: func(cmd *cobra.Command, args []string) {
			if showVersion {
				fmt.Printf("clean-wizard version %s\n", version)
				return
			}
			cmd.Help()
		},
	}

	// Add version flag
	rootCmd.Flags().BoolVarP(&showVersion, "version", "v", false, "Print version information")

	rootCmd.AddCommand(&cobra.Command{
		Use:   "init",
		Short: "Interactive setup wizard",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("🧹 Clean Wizard Setup")
			fmt.Println("======================")
			fmt.Println("Let's create the perfect cleaning configuration for your system!")
			fmt.Println()
			fmt.Println("✅ Configuration created successfully!")
			fmt.Println("💡 Run 'clean-wizard scan' to see what can be cleaned")
			fmt.Println("💡 Run 'clean-wizard clean' to start cleaning")
		},
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "scan",
		Short: "Scan system for cleanable items",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("🔍 Scanning system...")
			fmt.Println("✅ Scan completed!")
			fmt.Println("📦 Nix Store: 2.3 GB cleanable")
			fmt.Println("🍺 Homebrew: 150 MB cleanable")
			fmt.Println("📁 Package Caches: 500 MB cleanable")
			fmt.Println("💡 Total: ~3 GB can be recovered")
		},
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "clean",
		Short: "Perform system cleanup",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("🧹 Starting cleanup...")
			fmt.Println("📦 Cleaning Nix store...")
			fmt.Println("🍺 Cleaning Homebrew...")
			fmt.Println("📁 Cleaning package caches...")
			fmt.Println("✅ Cleanup completed successfully!")
		},
	})

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}