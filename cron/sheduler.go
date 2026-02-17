package cron

import (
	"github.com/robfig/cron/v3"
	"log"
	"magento.GO/config"
)

func StartCron() *cron.Cron {
	c := cron.New()
	for name, cronJob := range config.CronJobs {
		// Use the schedule and job function from the struct
		jobFunc := cronJob.Job
		_, err := c.AddFunc(cronJob.Schedule, func() { jobFunc() })
		if err != nil {
			log.Fatalf("Failed to register job %s: %v", name, err)
		}
	}
	c.Start()
	return c
}
