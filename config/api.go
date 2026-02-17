package config

// GetAuthSkipperPaths returns a list of paths to skip authentication for
func GetAuthSkipperPaths() []string {
	//Public API paths
	return []string{"/api/products", "/api/products/:id"}
}
