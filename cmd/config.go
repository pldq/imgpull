package cmd

import (
	"fmt"

	"image-poller/internal/config"

	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configure imgpull settings",
	Long:  `Configure Docker connection mode and other settings.`,
}

var configDockerCmd = &cobra.Command{
	Use:   "docker",
	Short: "Configure Docker connection",
	Long: `Configure how imgpull connects to Docker.

Examples:
  imgpull config docker --cli         # Use docker CLI (default)
  imgpull config docker --api         # Use Docker API with default socket
  imgpull config docker --api --host tcp://localhost:2375  # Use Docker API with custom host`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cliMode, _ := cmd.Flags().GetBool("cli")
		apiMode, _ := cmd.Flags().GetBool("api")
		host, _ := cmd.Flags().GetString("host")

		if cliMode && apiMode {
			return fmt.Errorf("cannot use both --cli and --api")
		}

		if cliMode {
			cfg.DockerMode = config.DockerModeCLI
			cfg.DockerHost = ""
		} else if apiMode {
			cfg.DockerMode = config.DockerModeAPI
			if host != "" {
				cfg.DockerHost = host
			}
		}

		dbPath, err := config.GetDBPath()
		if err != nil {
			return fmt.Errorf("failed to get database path: %w", err)
		}

		if err := cfg.Save(dbPath); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		fmt.Printf("Docker mode: %s\n", cfg.DockerMode)
		if cfg.DockerHost != "" {
			fmt.Printf("Docker host: %s\n", cfg.DockerHost)
		}
		return nil
	},
}

func init() {
	configDockerCmd.Flags().Bool("cli", false, "Use docker CLI")
	configDockerCmd.Flags().Bool("api", false, "Use Docker API")
	configDockerCmd.Flags().String("host", "", "Docker API host (e.g., tcp://localhost:2375)")
	configCmd.AddCommand(configDockerCmd)
	rootCmd.AddCommand(configCmd)
}
