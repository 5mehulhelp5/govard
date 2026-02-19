package cmd

import (
	"context"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "List all running Govard environments",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
		if err != nil {
			pterm.Error.Printf("Failed to connect to Docker: %v\n", err)
			return
		}

		containers, err := cli.ContainerList(ctx, container.ListOptions{})
		if err != nil {
			pterm.Error.Printf("Failed to list containers: %v\n", err)
			return
		}

		projects := make(map[string][]string)
		for _, c := range containers {
			for _, name := range c.Names {
				cleanName := strings.TrimPrefix(name, "/")
				// Standard Govard naming pattern: projectname-service-1
				parts := strings.Split(cleanName, "-")
				if len(parts) >= 3 {
					projectName := strings.Join(parts[:len(parts)-2], "-")
					projects[projectName] = append(projects[projectName], parts[len(parts)-2])
				}
			}
		}

		if len(projects) == 0 {
			pterm.Info.Println("No running Govard environments found.")
			return
		}

		pterm.DefaultHeader.WithFullWidth().Println("Govard Environments")

		tableData := pterm.TableData{
			{"Project", "Status", "Services"},
		}

		for project, services := range projects {
			// Basic filtering - only show if it looks like a govard project (has web/php/db)
			isGovard := false
			for _, s := range services {
				if s == "web" || s == "php" || s == "db" {
					isGovard = true
					break
				}
			}

			if isGovard {
				tableData = append(tableData, []string{
					pterm.Magenta(project),
					pterm.Green("Running"),
					strings.Join(services, ", "),
				})
			}
		}

		pterm.DefaultTable.WithHasHeader().WithData(tableData).Render()
	},
}
