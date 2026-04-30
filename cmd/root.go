package cmd

import (
	"fmt"

	"image-poller/internal/config"
	"image-poller/internal/db"

	"github.com/spf13/cobra"
)

var (
	cfg      *config.Config
	database *db.DB
)

var rootCmd = &cobra.Command{
	Use:   "imgpull",
	Short: "Pull Docker images via GitHub Actions proxy",
	Long: `imgpull is a tool that helps you pull Docker images from Docker Hub
through a GitHub Actions proxy. It creates branches in a GitHub repository,
triggers workflows to pull and push images to ghcr.io, then pulls them locally.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		dbPath, err := config.GetDBPath()
		if err != nil {
			return fmt.Errorf("failed to get database path: %w", err)
		}

		cfg, err = config.Load(dbPath)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		database, err = db.Open(dbPath)
		if err != nil {
			return fmt.Errorf("failed to open database: %w", err)
		}

		return nil
	},
	PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
		if database != nil {
			return database.Close()
		}
		return nil
	},
}

func Execute() {
	_ = rootCmd.Execute()
}
