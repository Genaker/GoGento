package html

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"html/template"
	"net/http"
	"time"
)

var (
	TemplateCompileTime time.Duration
	Templates           *template.Template
)

func initTemplates() {
	start := time.Now()
	tmpl := template.Must(template.New("hello_world.html").Funcs(template.FuncMap{
		"dict": Dict,
	}).ParseFiles("html/parts/hello_world.html"))
	compileTime := time.Since(start)
	TemplateCompileTime = compileTime
	Templates = tmpl
}

// RegisterHelloWorldRoute registers the /hello-world route
func RegisterHelloWorldRoute(e *echo.Echo) {
	initTemplates()
	e.GET("/hello-world", func(c echo.Context) error {

		// Retrieve the request registry from context
		reqRegIface := c.Get("RequestRegistry")
		var start time.Time
		var showTime bool
		if reqRegIface != nil {
			if reqReg, ok := reqRegIface.(interface {
				Get(string) (interface{}, bool)
			}); ok {
				if v, found := reqReg.Get("request_start"); found {
					if t, ok := v.(time.Time); ok {
						start = t
						showTime = true
					}
				}
			}
		}
		var dur time.Duration
		var execTime, execTimeMs string
		if showTime {
			dur = time.Since(start)
			execTime = dur.String()
			execTimeMs = fmt.Sprintf("%.4f ms", float64(dur.Nanoseconds())/1e6)
		} else {
			execTime = ""
			execTimeMs = ""
		}
		data := map[string]interface{}{
			"Message":               "Hello World",
			"ExecutionTime":         execTime,
			"ExecutionTimeMs":       execTimeMs,
			"TemplateCompileTime":   TemplateCompileTime.String(),
			"TemplateCompileTimeMs": fmt.Sprintf("%.10f ms", float64(TemplateCompileTime.Nanoseconds())/1e6),
		}
		return c.Render(http.StatusOK, "hello_world.html", data)
	})
}
