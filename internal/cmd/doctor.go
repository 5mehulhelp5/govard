package cmd

import (
	"encoding/json"
	"fmt"
	"govard/internal/engine"
	"os"
	"strings"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Run system diagnostics",
	Long: strings.TrimSpace(`
Run system diagnostics and report on the health of your local Govard environment.

Checks:
  - Docker daemon connectivity
  - Docker Compose plugin availability
  - Port conflicts on host (80/443)
  - Disk scratch write sanity
  - Govard home directory readiness (~/.govard)
  - Outbound network probe sanity
  - SSH agent connectivity and loaded keys

Use --fix to apply safe automatic remediations. Use --json for machine-readable output.
Use --pack to export a diagnostics support bundle for sharing with support.
`),
	Example: `  # Run a standard diagnostic pass
  govard doctor

  # Apply safe automatic fixes when available
  govard doctor --fix

  # Export a diagnostics support pack
  govard doctor --pack

  # Trust the Govard local CA
  govard doctor trust`,
	RunE: func(cmd *cobra.Command, args []string) error {
		outputJSON, _ := cmd.Flags().GetBool("json")
		fixEnabled, _ := cmd.Flags().GetBool("fix")
		packEnabled, _ := cmd.Flags().GetBool("pack")
		packDir, _ := cmd.Flags().GetString("pack-dir")
		if !outputJSON {
			fmt.Println()
			pterm.NewStyle(pterm.BgLightBlue, pterm.FgBlack, pterm.Bold).Println(" Govard System Doctor ")
			fmt.Println()
		}

		report := runDoctorDiagnostics()
		if fixEnabled {
			fixResults := applyDoctorSafeFixes(report)
			if outputJSON {
				for _, line := range summarizeDoctorFixResults(fixResults) {
					fmt.Fprintln(cmd.ErrOrStderr(), line)
				}
			} else {
				renderDoctorFixResults(fixResults)
			}
			report = runDoctorDiagnostics()
		}

		packPath := ""
		if packEnabled {
			cwd, _ := os.Getwd()
			resolvedPath, err := CreateDoctorDiagnosticsPack(packDir, cwd, report)
			if err != nil {
				return fmt.Errorf("create doctor diagnostics pack: %w", err)
			}
			packPath = resolvedPath
		}

		if outputJSON {
			payload, err := json.MarshalIndent(report, "", "  ")
			if err != nil {
				return fmt.Errorf("marshal doctor report: %w", err)
			}
			fmt.Fprintln(cmd.OutOrStdout(), string(payload))
		} else {
			renderDoctorReport(report)
		}
		if packPath != "" {
			if outputJSON {
				fmt.Fprintf(cmd.ErrOrStderr(), "Doctor diagnostics pack: %s\n", packPath)
			} else {
				pterm.Success.Printf("Doctor diagnostics pack: %s\n", packPath)
			}
		}

		if report.HasFailures() {
			return fmt.Errorf("doctor found %d blocking issue(s)", report.Failures)
		}
		return nil
	},
}

var doctorTrustCmd = &cobra.Command{
	Use:   "trust",
	Short: "Trust the local CA for SSL certificates",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if trustCmd.RunE == nil {
			return fmt.Errorf("doctor trust is unavailable")
		}
		return trustCmd.RunE(cmd, args)
	},
}

var doctorFixDepsCmd = &cobra.Command{
	Use:   "fix-deps",
	Short: "Check and report required system dependencies",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if fixDepsCmd.RunE == nil {
			return fmt.Errorf("doctor fix-deps is unavailable")
		}
		return fixDepsCmd.RunE(cmd, args)
	},
}

func renderDoctorReport(report engine.DoctorReport) {
	for _, check := range report.Checks {
		line := fmt.Sprintf("%s: %s", check.Title, check.Message)
		switch check.Status {
		case engine.DoctorStatusPass:
			pterm.Success.Println(line)
		case engine.DoctorStatusWarn:
			pterm.Warning.Println(line)
		case engine.DoctorStatusFail:
			pterm.Error.Println(line)
		default:
			pterm.Info.Println(line)
		}

		if check.Hint != "" {
			pterm.Info.Printf("Hint: %s\n", check.Hint)
		}
		if check.SuggestedCommand != "" {
			pterm.Info.Printf("Suggested next command: %s\n", check.SuggestedCommand)
		}
	}
	pterm.Info.Printf(
		"Summary: passed=%d warnings=%d failures=%d\n",
		report.Passed,
		report.Warnings,
		report.Failures,
	)
}

// DoctorCommand exposes the doctor command for tests.
func DoctorCommand() *cobra.Command {
	return doctorCmd
}

func init() {
	doctorCmd.Flags().Bool("json", false, "Print diagnostics as JSON")
	doctorCmd.Flags().Bool("fix", false, "Apply safe automatic fixes when available")
	doctorCmd.Flags().Bool("pack", false, "Export a diagnostics support pack")
	doctorCmd.Flags().String("pack-dir", "", "Output directory for diagnostics pack (default: ~/.govard/diagnostics)")
	doctorCmd.AddCommand(doctorTrustCmd)
	doctorCmd.AddCommand(doctorFixDepsCmd)
}
