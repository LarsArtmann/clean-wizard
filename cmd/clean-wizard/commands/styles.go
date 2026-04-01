package commands

import (
	"fmt"

	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/table"
	"github.com/spf13/cobra"
)

var (
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("86")).
			Padding(1, 0)

	WarningStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")).
			Padding(0, 1)

	InfoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("117")).
			Padding(0, 1)

	SuccessStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("82")).
			Padding(0, 1)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Padding(0, 1)

	HeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("81")).
			Padding(0, 1)

	MutedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Padding(0, 1)
)

const (
	ProfileFormatText  = "text"
	ProfileFormatJSON  = "json"
	ProfileFormatEmoji = "emoji"
)

func newResultsTable(rows ...[]string) *table.Table {
	return table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("238"))).
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == 0 {
				return lipgloss.NewStyle().Foreground(lipgloss.Color("86")).Bold(true)
			}
			return lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
		}).
		Rows(rows...)
}

func newParentCommand(
	name, shortDesc, longDesc string,
	subcommands ...func() *cobra.Command,
) *cobra.Command {
	cmd := &cobra.Command{
		Use:   name,
		Short: shortDesc,
		Long:  longDesc,
	}
	for _, subCmd := range subcommands {
		cmd.AddCommand(subCmd())
	}
	return cmd
}

func PrintProfileSummaries(profiles any, format string) {
	switch format {
	case ProfileFormatJSON:
		fmt.Printf("%v\n", profiles)
	case ProfileFormatEmoji:
		fmt.Printf("📋 Profiles: %v\n", profiles)
	default:
		fmt.Printf("Profiles:\n%v\n", profiles)
	}
}
