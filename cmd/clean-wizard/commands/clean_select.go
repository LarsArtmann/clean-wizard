package commands

import (
	"fmt"

	"charm.land/huh/v2"
	"github.com/LarsArtmann/clean-wizard/internal/domain"
	errorfamily "github.com/larsartmann/go-error-family"
)

// Classified sentinel errors for clean_select.
var (
	ErrProfileNotFound   = errorfamily.NewRejection("clean.profile_not_found", "profile not found")
	ErrProfileNoCleaners = errorfamily.NewRejection("clean.profile_no_cleaners", "profile has no available cleaners")
)

// destructiveCleaners are excluded from "standard" mode because they are
// potentially destructive or require external services (e.g., Docker daemon).
var destructiveCleaners = map[CleanerType]bool{ //nolint:gochecknoglobals
	CleanerTypeDocker:                       true,
	CleanerTypeProjectsManagementAutomation: true,
}

// selectCleaners determines which cleaners to run based on priority:
// profile > mode > interactive.
func selectCleaners(
	profile, mode string,
	cfg *domain.Config,
	availableConfigs []CleanerConfig,
	jsonOutput bool,
) ([]CleanerType, error) {
	switch {
	case profile != "":
		return selectProfileCleaners(profile, cfg, availableConfigs, jsonOutput)
	case mode != "":
		return selectPresetCleaners(mode, availableConfigs, jsonOutput)
	case jsonOutput:
		return allAvailableTypes(availableConfigs), nil
	default:
		return selectCleanersInteractive(availableConfigs)
	}
}

func selectProfileCleaners(
	profileName string,
	cfg *domain.Config,
	availableConfigs []CleanerConfig,
	jsonOutput bool,
) ([]CleanerType, error) {
	selected, err := getProfileCleaners(profileName, cfg, availableConfigs)
	if err != nil {
		return nil, fmt.Errorf("profile error for profileName=%v: %w", profileName, err)
	}

	if !jsonOutput {
		fmt.Printf("📋 Using profile: %s\n\n", profileName)
		printSelectedCleaners(selected)
	}

	return selected, nil
}

func selectPresetCleaners(
	mode string,
	availableConfigs []CleanerConfig,
	jsonOutput bool,
) ([]CleanerType, error) {
	selected := getPresetSelection(mode, availableConfigs)

	if !jsonOutput {
		fmt.Printf("🎯 Using preset mode: %s\n\n", mode)
		printSelectedCleaners(selected)
	}

	return selected, nil
}

func printSelectedCleaners(cleaners []CleanerType) {
	for _, ct := range cleaners {
		fmt.Printf("  ✓ %s\n", getCleanerName(ct))
	}

	fmt.Println()
}

func selectCleanersInteractive(availableConfigs []CleanerConfig) ([]CleanerType, error) {
	fmt.Println(InfoStyle.Render("⌨️  Keyboard Shortcuts:"))
	fmt.Println("   ↑↓ : Navigate  |  Space : Select  |  Enter : Confirm  |  Esc : Cancel")
	fmt.Println()

	var selectedTypes []CleanerType

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[CleanerType]().
				Title("Select cleaners to run").
				Description("Choose which cleaners to execute (Space to select, Enter to confirm)").
				Options(
					func() []huh.Option[CleanerType] {
						opts := make([]huh.Option[CleanerType], len(availableConfigs))
						for i, cfg := range availableConfigs {
							desc := fmt.Sprintf("%s %s", cfg.Description, cfg.Icon)
							opts[i] = huh.NewOption(desc, cfg.Type)
						}

						return opts
					}()...,
				).
				Value(&selectedTypes),
		),
	)

	err := form.Run()
	if err != nil {
		return nil, errorfamily.WrapRejection(err, "clean.select_form", "form error")
	}

	if len(selectedTypes) == 0 {
		return nil, nil
	}

	return selectedTypes, nil
}

func confirmExecution(skipConfirmation, dryRun bool) (bool, error) {
	if dryRun || skipConfirmation {
		return true, nil
	}

	var confirm bool

	confirmForm := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Run selected cleaners?").
				Affirmative("Yes, run them").
				Negative("No, cancel").
				Value(&confirm),
		),
	)

	err := confirmForm.Run()
	if err != nil {
		return false, errorfamily.WrapRejection(err, "clean.confirm", "confirmation error")
	}

	return confirm, nil
}

// getPresetSelection returns cleaner selection based on preset mode.
func getPresetSelection(mode string, configs []CleanerConfig) []CleanerType {
	switch mode {
	case "quick":
		return []CleanerType{
			CleanerTypeHomebrew,
			CleanerTypeNodePackages,
			CleanerTypeGoPackages,
			CleanerTypeTempFiles,
			CleanerTypeBuildCache,
		}
	case "aggressive": //nolint:goconst
		return allAvailableTypes(configs)
	case "standard":
		var safeTypes []CleanerType

		for _, cfg := range configs {
			if !destructiveCleaners[cfg.Type] {
				safeTypes = append(safeTypes, cfg.Type)
			}
		}

		return safeTypes
	default:
		return allAvailableTypes(configs)
	}
}

func allAvailableTypes(configs []CleanerConfig) []CleanerType {
	types := make([]CleanerType, len(configs))
	for i, cfg := range configs {
		types[i] = cfg.Type
	}

	return types
}

// getProfileCleaners returns the cleaner types for a given profile name.
func getProfileCleaners(
	profileName string,
	cfg *domain.Config,
	availableConfigs []CleanerConfig,
) ([]CleanerType, error) {
	profile, exists := cfg.Profiles[profileName]
	if !exists {
		return nil, fmt.Errorf("%w: %q", ErrProfileNotFound, profileName)
	}

	availableSet := make(map[CleanerType]bool)
	for _, ac := range availableConfigs {
		availableSet[ac.Type] = true
	}

	var cleaners []CleanerType

	for _, op := range profile.Operations {
		if op.Enabled != domain.ProfileStatusEnabled {
			continue
		}

		opType := domain.GetOperationType(op.Name)

		cleanerType, ok := operationTypeToCleanerType[opType]
		if !ok {
			continue
		}

		if availableSet[cleanerType] {
			cleaners = append(cleaners, cleanerType)
		}
	}

	if len(cleaners) == 0 {
		return nil, fmt.Errorf("%w: %q", ErrProfileNoCleaners, profileName)
	}

	return cleaners, nil
}
