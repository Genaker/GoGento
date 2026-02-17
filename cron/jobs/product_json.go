package jobs

import (
	"fmt"
	"time"
)

func ProductJsonJob(params ...string) {
	fmt.Println("Running ProductJsonJob at", time.Now())
	fmt.Println("Params:", params)
	// Your job logic here
}
