package parts

import (
	"log"
	"os"
	"sync"
)

var (
	criticalCSS     string
	criticalCSSErr  error
	criticalCSSOnce sync.Once
)

func GetCriticalCSSCached() (string, error) {
	criticalCSSOnce.Do(func() {
		criticalCSS, criticalCSSErr = GetCriticalCSS()
	})
	return criticalCSS, criticalCSSErr
}

// GetCriticalCSS reads the critical CSS file and returns it as a string.
func GetCriticalCSS() (string, error) {
	css, err := os.ReadFile("assets/tailwind.min.css")
	if err != nil {
		log.Println("Critical CSS error:", err)
		return "", err
	}
	return string(css), nil
}
