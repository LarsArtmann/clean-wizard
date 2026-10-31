package commands

import (
	"context"
	"encoding/json/jsontext"
	"encoding/json/v2"
	"fmt"
	"os"

	"github.com/LarsArtmann/clean-wizard/internal/di"
	"github.com/LarsArtmann/clean-wizard/internal/format"
	errorfamily "github.com/larsartmann/go-error-family"
	"github.com/samber/do/v2"
	"github.com/spf13/cobra"
)

// ErrDoctorUnhealthy is the classified Infrastructure sentinel for a failed
// doctor run: errorfamily.ExitCode maps it to sysexits 69 at the CLI boundary.
var ErrDoctorUnhealthy = errorfamily.NewInfrastructure(
	"doctor.unhealthy",
	"one or more DI services failed construction or health check",
)

// ErrDoctorJSONWithSARIF rejects --json combined with --sarif: both are
// machine channels and their outputs would silently disagree about what the
// run "output" is.
var ErrDoctorJSONWithSARIF = errorfamily.NewRejection(
	"doctor.json_sarif_conflict",
	"--json and --sarif are mutually exclusive; both are machine channels",
)

// NewDoctorCommand creates the health-check command: materialize every DI
// service, run container health checks, render per-service outcomes.
func NewDoctorCommand() *cobra.Command {
	var (
		verbose    bool
		configPath string
		jsonOutput bool
		sarifPath  string
	)

	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Check the health of clean-wizard's DI services",
		Long: `Materialize every DI service and run the container health check.

Prints one line per service: construction plus health-check outcome. Exit code
is 0 when every service is healthy, 69 (Infrastructure) when any service
failed — suitable for scripting and CI gates.

Machine channels: --json prints a JSON report to stdout; --sarif writes a
SARIF 2.1.0 document (error findings for unhealthy services only) to a file
path or "-" for stdout.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDoctorCommand(doctorOptions{
				verbose:    verbose,
				configPath: configPath,
				jsonOutput: jsonOutput,
				sarifPath:  sarifPath,
			})
		},
	}

	cmd.Flags().BoolVar(&verbose, "verbose", false, "Enable verbose output")
	cmd.Flags().StringVarP(&configPath, "config", "c", "", "Path to configuration file")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output health report in JSON format")
	cmd.Flags().StringVar(&sarifPath, "sarif", "",
		"Write health findings as SARIF to this path (or '-' for stdout; off with --json)")

	return cmd
}

// doctorOptions carries the doctor invocation flags.
type doctorOptions struct {
	verbose    bool
	configPath string
	jsonOutput bool
	sarifPath  string
}

// runDoctorCommand builds the container exactly like clean/scan (config load →
// di.New → RegisterAllServices), reports health, and maps failures to a
// classified Infrastructure error.
func runDoctorCommand(opts doctorOptions) error {
	if opts.jsonOutput && opts.sarifPath != "" {
		return ErrDoctorJSONWithSARIF
	}

	ctx := context.Background()

	cfg, err := loadConfigFromPath(opts.configPath)
	if err != nil {
		return errorfamily.WrapRejection(err, "doctor.config_load", "failed to load configuration")
	}

	container, cleanup := di.New()
	defer cleanup()

	settings := di.RunSettings{
		Verbose:        opts.verbose,
		DryRun:         false,
		MaxConcurrency: 0,
		Profile:        "",
	}
	if err := di.RegisterAllServices(container.Injector(), cfg, settings); err != nil {
		return errorfamily.WrapRejection(err, "doctor.di_register", "failed to register DI services")
	}

	report, err := di.HealthReport(ctx, container.Injector())
	if err != nil {
		return errorfamily.WrapInfrastructure(err, "doctor.health_report", "failed to run container health checks")
	}

	switch {
	case opts.jsonOutput:
		if err := outputDoctorJSON(report); err != nil {
			return err
		}
	case opts.sarifPath != "":
		if err := writeDoctorSARIF(report, opts.sarifPath); err != nil {
			return err
		}
	default:
		printDoctorReport(report)
	}

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

// enforceCleanHealthGate aborts a clean run when the container is unhealthy.
// Opt-in via --check-health; failures are classified Infrastructure so the
// CLI exits 69 instead of half-executing a clean with broken services.
func enforceCleanHealthGate(ctx context.Context, injector do.Injector) error {
	report, err := di.HealthReport(ctx, injector)
	if err != nil {
		return errorfamily.WrapInfrastructure(err, "clean.health_report", "failed to run pre-run health checks")
	}

	summary := di.Summarize(report)
	if summary.Unhealthy == 0 {
		return nil
	}

	return fmt.Errorf("refusing to clean with unhealthy services: %w", di.UnhealthyError(summary))
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

// healthJSONOutput is the doctor --json document shape. Field enrichment
// (family/code/retryable) mirrors the clean/scan JSON schemas.
type healthJSONService struct {
	Service   string `json:"service"`
	Healthy   bool   `json:"healthy"`
	Error     string `json:"error,omitempty"`
	Family    string `json:"family,omitempty"`
	Code      string `json:"code,omitempty"`
	Retryable bool   `json:"retryable,omitempty"`
}

type healthJSONSummary struct {
	Total     int  `json:"total"`
	Unhealthy int  `json:"unhealthy"`
	Healthy   bool `json:"healthy"`
}

type healthJSONOutput struct {
	Services []healthJSONService `json:"services"`
	Summary  healthJSONSummary   `json:"summary"`
}

// outputDoctorJSON writes the health report as JSON.
func outputDoctorJSON(report []di.HealthEntry) error {
	jsonBytes, err := buildDoctorJSON(report)
	if err != nil {
		return err
	}

	fmt.Println(string(jsonBytes))

	return nil
}

// buildDoctorJSON renders the health report as indented JSON bytes.
func buildDoctorJSON(report []di.HealthEntry) ([]byte, error) {
	jsonServices := make([]healthJSONService, 0, len(report))
	for _, entry := range report {
		service := healthJSONService{
			Service: entry.Service,
			Healthy: entry.Healthy(),
		}

		if entry.Error != nil {
			family := errorfamily.Classify(entry.Error)
			service.Error = entry.Error.Error()
			service.Family = family.String()
			service.Code = errorfamily.Code(entry.Error)
			service.Retryable = family.IsRetryable()
		}

		jsonServices = append(jsonServices, service)
	}

	summary := di.Summarize(report)
	output := healthJSONOutput{
		Services: jsonServices,
		Summary: healthJSONSummary{
			Total:     summary.Total,
			Unhealthy: summary.Unhealthy,
			Healthy:   summary.Unhealthy == 0,
		},
	}

	jsonBytes, err := json.Marshal(output, jsontext.WithIndentPrefix(""), jsontext.WithIndent("  "))
	if err != nil {
		return nil, errorfamily.WrapCorruption(err, "doctor.json_output", "failed to generate doctor JSON output")
	}

	return jsonBytes, nil
}

// writeDoctorSARIF writes health findings as a SARIF document to a path (or
// stdout for "-"). Healthy services produce no finding.
func writeDoctorSARIF(report []di.HealthEntry, sarifPath string) error {
	outcomes := make([]format.HealthOutcome, 0, len(report))
	for _, entry := range report {
		outcomes = append(outcomes, format.HealthOutcome{Service: entry.Service, Err: entry.Error})
	}

	data, err := format.HealthOutcomesToSARIF(outcomes)
	if err != nil {
		return err
	}

	if sarifPath == "-" {
		fmt.Println(string(data))

		return nil
	}

	if writeErr := os.WriteFile(sarifPath, data, 0o600); writeErr != nil { //nolint:mnd
		return errorfamily.WrapInfrastructuref(
			writeErr,
			"doctor.sarif_write",
			"failed to write SARIF report to %s",
			sarifPath,
		)
	}

	fmt.Printf("SARIF report written to %s\n", sarifPath)

	return nil
}
