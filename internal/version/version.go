package version

// These variables are set at build time using ldflags
var (
	// Version is the semantic version (e.g., "0.1.0")
	Version = "dev"

	// GitCommit is the git commit hash
	GitCommit = "unknown"

	// BuildDate is the build timestamp
	BuildDate = "unknown"
)

// Info returns a formatted version string
func Info() string {
	return Version
}

// Full returns the full version info including commit and date
func Full() string {
	return Version + " (commit: " + GitCommit + ", built: " + BuildDate + ")"
}
