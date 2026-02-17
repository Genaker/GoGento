package html

import (
	"github.com/labstack/echo/v4"
	"html/template"
	"io"
)

// Template is the HTML template renderer
type Template struct {
	Templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.Templates.ExecuteTemplate(w, name, data)
}

// Dict creates a map from pairs of values for use in templates
func Dict(values ...interface{}) map[string]interface{} {
	if len(values)%2 != 0 {
		panic("dict expects an even number of arguments")
	}
	m := make(map[string]interface{}, len(values)/2)
	for i := 0; i < len(values); i += 2 {
		key, ok := values[i].(string)
		if !ok {
			panic("dict keys must be strings")
		}
		m[key] = values[i+1]
	}
	return m
}

// Template functions for all templates
var templateFuncs = template.FuncMap{
	// Math operations
	"add": func(a, b int) int { return a + b },
	"sub": func(a, b int) int { return a - b },
	"mul": func(a, b int) int { return a * b },
	"div": func(a, b int) int {
		if b == 0 {
			return 0
		}
		return a / b
	},

	// Comparison helpers
	"eq": func(a, b interface{}) bool { return a == b },
	"ne": func(a, b interface{}) bool { return a != b },
	"lt": func(a, b int) bool { return a < b },
	"gt": func(a, b int) bool { return a > b },
	"le": func(a, b int) bool { return a <= b },
	"ge": func(a, b int) bool { return a >= b },

	// Slice helpers
	"until": func(count int) []int {
		s := make([]int, count)
		for i := 0; i < count; i++ {
			s[i] = i + 1
		}
		return s
	},
	"range_offset": func(start, end int) []int {
		n := end - start + 1
		if n <= 0 {
			return []int{}
		}
		r := make([]int, n)
		for i := 0; i < n; i++ {
			r[i] = start + i
		}
		return r
	},

	// Map helper
	"dict": func(values ...interface{}) map[string]interface{} {
		if len(values)%2 != 0 {
			return nil
		}
		dict := make(map[string]interface{}, len(values)/2)
		for i := 0; i < len(values); i += 2 {
			key, ok := values[i].(string)
			if !ok {
				continue
			}
			dict[key] = values[i+1]
		}
		return dict
	},

	// String operations
	"concat": func(str ...string) string {
		result := ""
		for _, s := range str {
			result += s
		}
		return result
	},
}

// GetTemplateFuncs returns the template functions map
func GetTemplateFuncs() template.FuncMap {
	return templateFuncs
}

// NewTemplate creates a new template with all functions registered
func NewTemplate() *Template {
	return &Template{
		Templates: template.Must(template.New("").Funcs(templateFuncs).ParseGlob("html/parts/*.html")),
	}
}
