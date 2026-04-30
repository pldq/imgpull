package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List pulled images",
	Long:  `List all images that have been pulled using imgpull.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		limit, _ := cmd.Flags().GetInt("limit")

		records, err := database.ListRecords(limit)
		if err != nil {
			return fmt.Errorf("failed to list records: %w", err)
		}

		if len(records) == 0 {
			fmt.Println("No images pulled yet.")
			return nil
		}

		fmt.Printf("%-5s %-30s %-12s %-12s %-10s %-10s %-20s %-20s\n", "ID", "Image", "Tag", "Size", "Duration", "First", "FirstPullTime", "PullTime")
		fmt.Println(strings.Repeat("-", 130))

		for _, r := range records {
			sizeStr := formatSize(r.Size)
			durationStr := time.Duration(r.Duration * int64(time.Millisecond)).Round(time.Second).String()
			firstDurationStr := time.Duration(r.FirstDuration * int64(time.Millisecond)).Round(time.Second).String()
			imageName := r.ImageName
			if len(imageName) > 28 {
				imageName = imageName[:25] + "..."
			}
			firstPullTimeStr := r.FirstPullTime.Format("2006-01-02 15:04:05")
			pullTimeStr := r.PullTime.Format("2006-01-02 15:04:05")
			fmt.Printf("%-5d %-30s %-12s %-12s %-10s %-10s %-20s %-20s\n", r.ID, imageName, r.Tag, sizeStr, durationStr, firstDurationStr, firstPullTimeStr, pullTimeStr)
		}

		return nil
	},
}

func formatSize(size int64) string {
	const KB = 1024
	const MB = KB * 1024
	const GB = MB * 1024

	if size >= GB {
		return fmt.Sprintf("%.2f GB", float64(size)/GB)
	}
	if size >= MB {
		return fmt.Sprintf("%.2f MB", float64(size)/MB)
	}
	if size >= KB {
		return fmt.Sprintf("%.2f KB", float64(size)/KB)
	}
	return fmt.Sprintf("%d B", size)
}

func init() {
	listCmd.Flags().IntP("limit", "n", 20, "Limit number of results")
	rootCmd.AddCommand(listCmd)
}
