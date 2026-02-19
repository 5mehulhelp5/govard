package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"govard/internal/engine"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var snapshotCmd = &cobra.Command{
	Use:   "snapshot",
	Short: "Manage local snapshots for database and media",
}

var snapshotCreateCmd = &cobra.Command{
	Use:   "create [name]",
	Short: "Create a snapshot of local database and media",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		config := loadFullConfig()
		cwd, _ := os.Getwd()

		name := ""
		if len(args) == 1 {
			name = args[0]
		}

		path, err := engine.CreateSnapshot(cwd, config, name)
		if err != nil {
			pterm.Error.Printf("Snapshot create failed: %v\n", err)
			return
		}

		pterm.Success.Printf("Snapshot created at %s\n", path)
	},
}

var snapshotListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available snapshots",
	Run: func(cmd *cobra.Command, args []string) {
		cwd, _ := os.Getwd()
		snapshots, err := engine.ListSnapshots(cwd)
		if err != nil {
			pterm.Error.Printf("Snapshot list failed: %v\n", err)
			return
		}

		if len(snapshots) == 0 {
			pterm.Info.Println("No snapshots found.")
			return
		}

		w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 2, 2, ' ', 0)
		_, _ = fmt.Fprintln(w, "NAME\tCREATED_AT\tDB\tMEDIA")
		for _, snapshot := range snapshots {
			created := "-"
			if !snapshot.CreatedAt.IsZero() {
				created = snapshot.CreatedAt.Format("2006-01-02 15:04:05")
			}
			_, _ = fmt.Fprintf(w, "%s\t%s\t%t\t%t\n", snapshot.Name, created, snapshot.DB, snapshot.Media)
		}
		_ = w.Flush()
	},
}

var snapshotRestoreCmd = &cobra.Command{
	Use:   "restore <name>",
	Short: "Restore a snapshot",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		config := loadFullConfig()
		cwd, _ := os.Getwd()
		name := args[0]

		dbOnly, _ := cmd.Flags().GetBool("db-only")
		mediaOnly, _ := cmd.Flags().GetBool("media-only")
		if dbOnly && mediaOnly {
			pterm.Error.Println("Cannot use --db-only and --media-only together.")
			return
		}

		if err := engine.RestoreSnapshot(cwd, config, name, dbOnly, mediaOnly); err != nil {
			pterm.Error.Printf("Snapshot restore failed: %v\n", err)
			return
		}

		pterm.Success.Printf("Snapshot %s restored.\n", name)
	},
}

func init() {
	snapshotRestoreCmd.Flags().Bool("db-only", false, "Restore database only")
	snapshotRestoreCmd.Flags().Bool("media-only", false, "Restore media only")

	snapshotCmd.AddCommand(snapshotCreateCmd)
	snapshotCmd.AddCommand(snapshotListCmd)
	snapshotCmd.AddCommand(snapshotRestoreCmd)

	rootCmd.AddCommand(snapshotCmd)
}
