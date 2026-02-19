package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var shellCmd = &cobra.Command{
	Use:   "shell",
	Short: "Enter the application container",
	Run: func(cmd *cobra.Command, args []string) {
		config := loadConfig()
		containerName := fmt.Sprintf("%s-php-1", config.ProjectName)

		user := "www-data"

		c := exec.Command("docker", "exec", "-it", "-u", user, containerName, "bash")
		c.Stdin, c.Stdout, c.Stderr = os.Stdin, os.Stdout, os.Stderr

		// Try bash, fallback to sh if it fails
		if err := c.Run(); err != nil {
			c = exec.Command("docker", "exec", "-it", "-u", user, containerName, "sh")
			c.Stdin, c.Stdout, c.Stderr = os.Stdin, os.Stdout, os.Stderr
			c.Run()
		}
	},
}
