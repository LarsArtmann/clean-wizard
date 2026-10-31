package commands

import (
	"context"
	"fmt"

	"github.com/LarsArtmann/clean-wizard/internal/di"
	errorfamily "github.com/larsartmann/go-error-family"
	"github.com/spf13/cobra"
)

// ErrDoctorUnhealthy is the classified Infrastructure sentinel for a failed
// doctor run: errorfamily.ExitCode maps it to sysexits 69 at the CLI boundary.
var ErrDoctorUnhealthy = errorfamily.NewInfrastructure(
	"doctor.unhealthy",
	"one or more DI services failed construction or health check",
)

// NewDoctorCommand creates the health-check command: materialize every DI
// service, run container health checks, render per-service outcomes.
func NewDoctorCommand() *cobra.Command {
	var (
		verbose    bool
		configPath string
	)

	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Check the health of clean-wizard's DI services",
		Long: `Materialize every DI service and run the container health check.

Prints one line per service: construction plus health-check outcome. Exit code
is 0 when every service is healthy, 69 (Infrastructure) when any service
failed — suitable for scripting and CI gates.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDoctorCommand(verbose, configPath)
		},
	}

	cmd.Flags().BoolVar(&verbose, "verbose", false, "Enable verbose output")
	cmd.Flags().StringVarP(&configPath, "config", "c", "", "Path to configuration file")

	return cmd
}

// runDoctorCommand builds the container exactly like clean/scan (config load →
// di.New → RegisterAllServices), reports health, and maps failures to a
// classified Infrastructure error.
func runDoctorCommand(verbose bool, configPath string) error {
	ctx := context.Background()

	cfg, err := loadConfigFromPath(configPath)
	if err != nil {
		return errorfamily.WrapRejection(err, "doctor.config_load", "failed to load configuration")
	}

	container, cleanup := di.New()
	defer cleanup()

	settings := di.RunSettings{Verbose: verbose, DryRun: false, MaxConcurrency: 0, Profile: ""} //nolint:exhaustruct // doctor reads no config profile
	if err := di.RegisterAllServices(container.Injector(), cfg, settings); err != nil {
		return errorfamily.WrapRejection(err, "doctor.di_register", "failed to register DI services")
	}

	report, err := di.HealthReport(ctx, container.Injector())
	if err != nil {
		return errorfamily.WrapInfrastructure(err, "doctor.health_report", "failed to run container health checks")
	}

	printDoctorReport(report)

	summary := di.Summarize(report)
	if summary.Unhealthy > 0 {
		return fmt.Errorf("%w (%d of %d services)", ErrDoctorUnhealthy, summary.Unhealthy, summary.Total)
	}

	return nil
}

// printDoctorReport renders the human report: one ✅/❌ line per service plus a
// summary; unhealthy entries get their classified error detail.
func printDoctorReport(report []di.HealthEntry) {
	fmt.Println(TitleStyle.Render("🩺 Clean Wizard Doctor"))
	fmt.Println()

	for _, entry := range report {
		if entry.Healthy() {
			fmt.Printf("%s %s\n", SuccessStyle.Render("✅"), shortServiceName(entry.Service))

			continue
		}

		fmt.Printf("%s %s\n", ErrorStyle.Render("❌"), shortServiceName(entry.Service))
		fmt.Printf("   %v\n", entry.Error)
	}

	fmt.Println()

	summary := di.Summarize(report)
	if summary.Unhealthy > 0 {
		fmt.Printf("❌ %d of %d service(s) unhealthy\n", summary.Unhealthy, summary.Total)

		return
	}

	fmt.Printf("✅ All %d service(s) healthy\n", summary.Total)
}

// shortServiceName trims the Go module path from a fully-qualified DI service
// name for display ("github.com/m/p/adapters.NixAdapter" → "adapters.NixAdapter").
func shortServiceName(serviceName string) string {
	for i := len(serviceName) - 1; i >= 0; i-- {
		if serviceName[i] == '/' {
			return serviceName[i+1:]
		}
	}

	return serviceName
}
