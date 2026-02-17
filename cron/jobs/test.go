package jobs

import (
	"fmt"
	"time"
)

func TestJob(params ...string) {
	elapsed, b := testGo()
	fmt.Printf("[TestJob] Loop completed in %.6f seconds. Last b=%d\n", elapsed, b)
	fmt.Println("Params:", params)
}

func testGo() (float64, int) {
	start := time.Now()
	var b int
	// Start of the code to profile
	for a := 0; a < 10000000; a++ {
		b = (a * a) // Use blank identifier to ignore result
	}
	// End of the code to profile
	time := time.Since(start).Seconds()
	return time, b
}
