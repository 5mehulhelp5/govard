package cmd

import (
	"fmt"

	"govard/internal/engine"
	"govard/internal/ui"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var projectCmd = &cobra.Command{
	Use:     "project",
	Aliases: []string{"prj", "projects", "registry"},
	Short:   "Browse known projects from registry",
}

var projectOpenCmd = &cobra.Command{
	Use:   "open <query>",
	Short: "Find a project by fuzzy query and print its path",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runProjectOpen(cmd, args[0])
	},
}

var projectListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all available projects in the registry",
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runProjectList(cmd)
	},
}

var projectDeleteForce bool

var projectDeleteCmd = &cobra.Command{
	Use:     "delete <query>",
	Aliases: []string{"rm", "del", "remove"},
	Short:   "Completely remove a project and its Docker resources",
	Long: `Permanently delete a project from Govard.
This command will:
1. Stop all containers for the project.
2. Remove all Docker volumes (database data, etc.).
3. Unregister project domains from proxy and hosts.
4. Remove the project from the Govard registry.

WARNING: This action is destructive and cannot be undone (for volumes).
It does NOT delete your project source code.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runProjectDelete(cmd, args[0])
	},
}

func initProjectCommands() {
	projectCmd.AddCommand(projectOpenCmd)
	projectCmd.AddCommand(projectListCmd)

	projectDeleteCmd.Flags().BoolVarP(&projectDeleteForce, "force", "f", false, "Delete without confirmation")
	projectCmd.AddCommand(projectDeleteCmd)
}

func runProjectOpen(cmd *cobra.Command, query string) error {
	match, err := engine.FindProjectByQuery(query)
	if err != nil {
		return err
	}

	_, _ = fmt.Fprintln(cmd.OutOrStdout(), match.Path)
	return nil
}

func runProjectDelete(cmd *cobra.Command, query string) error {
	match, err := engine.FindProjectByQuery(query)
	if err != nil {
		return err
	}

	if !projectDeleteForce {
		pterm.Warning.Printf("You are about to delete project: %s\n", match.ProjectName)
		pterm.Warning.Println("This will remove all Docker containers and VOLUMES (database data).")
		pterm.Warning.Printf("Project path: %s\n", match.Path)
		fmt.Println()

		result, _ := pterm.DefaultInteractiveConfirm.WithDefaultValue(false).Show("Are you sure you want to proceed?")
		if !result {
			pterm.Info.Println("Deletion cancelled.")
			return nil
		}
	}

	fmt.Println()
	pterm.NewStyle(pterm.BgLightBlue, pterm.FgBlack, pterm.Bold).Printf(" DELETING PROJECT: %s \n", match.ProjectName)
	fmt.Println()

	spinner, _ := pterm.DefaultSpinner.Start("Cleaning up resources...")
	err = engine.DeleteProject(cmd.Context(), match.Path, ui.NewPtermWriter(&pterm.Info), ui.NewPtermWriter(&pterm.Error))
	if err != nil {
		spinner.Fail(err.Error())
		return err
	}
	spinner.Success("Project deleted successfully.")

	return nil
}

func runProjectList(cmd *cobra.Command) error {
	entries, err := engine.ReadProjectRegistryEntries()
	if err != nil {
		return err
	}

	if len(entries) == 0 {
		pterm.Info.Println("The project registry is empty. Initialize projects with 'govard init'.")
		return nil
	}

	tableData := [][]string{
		{"Project", "Framework", "Domain", "Path"},
	}

	for _, entry := range entries {
		framework := entry.Framework
		if framework == "" {
			framework = "unknown"
		}
		tableData = append(tableData, []string{
			entry.ProjectName,
			framework,
			entry.Domain,
			entry.Path,
		})
	}

	err = pterm.DefaultTable.WithHasHeader().WithData(tableData).Render()
	if err != nil {
		return err
	}

	return nil
}
