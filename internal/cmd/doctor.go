package cmd

import (
	"encoding/json"
	"fmt"
	"govard/internal/engine"
	"os"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Run system diagnostics",
	RunE: func(cmd *cobra.Command, args []string) error {
		outputJSON, _ := cmd.Flags().GetBool("json")
		packEnabled, _ := cmd.Flags().GetBool("pack")
		packDir, _ := cmd.Flags().GetString("pack-dir")
		if !outputJSON {
			pterm.DefaultHeader.Println("Govard System Doctor")
		}

		report := engine.RunDoctorDiagnostics(engine.DoctorDependencies{})
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
	doctorCmd.Flags().Bool("pack", false, "Export a diagnostics support pack")
	doctorCmd.Flags().String("pack-dir", "", "Output directory for diagnostics pack (default: ~/.govard/diagnostics)")
}
