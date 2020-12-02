package env

var (
	// Version of app
	Version string
	// Build date for version
	Build string
)

// GetVersion returns app version
func GetVersion() string {
	return Version
}

// GetBuild returns app build date
func GetBuild() string {
	return Build
}
