package cmd

import (
	"fmt"
	"magento.GO/cron"
	"os"
	"strings"
	//"magento.GO/cron/jobs"
	"github.com/spf13/cobra"
	"magento.GO/config"
)

var jobName string

var cronStartCmd = &cobra.Command{
	Use:   "cron:start",
	Short: "Start the cron scheduler or run a single job by name",
	Run: func(cmd *cobra.Command, args []string) {
		if jobName != "" {
			cronJob, ok := config.CronJobs[strings.ToLower(jobName)]
			if !ok {
				fmt.Printf("Unknown job: %s\n", jobName)
				os.Exit(1)
			}
			fmt.Printf("Running single cron job: %s\n", jobName)
			cronJob.Job(args...)
			return
		}
		fmt.Println("Starting cron scheduler...")
		c := cron.StartCron()
		defer c.Stop()
		fmt.Println("Cron scheduler started. Press Ctrl+C to exit.")
		select {} // Block forever
	},
}

func init() {
	cronStartCmd.Flags().StringVarP(&jobName, "job", "j", "", "Run a single cron job by name and exit")
	rootCmd.AddCommand(cronStartCmd)
}
