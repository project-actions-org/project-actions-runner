package cli

// Version information
// These variables are set via ldflags during build
var (
	// Version is the semantic version
	Version = "0.3.3"

	// BuildTime is the build timestamp in format YYYYMMDDHHMM
	BuildTime = "dev"

	// FullVersion returns the complete version string
	FullVersion = Version + "+" + BuildTime
)

// GetVersion returns the full version string with build identifier
func GetVersion() string {
	if BuildTime == "dev" {
		return Version + "+dev"
	}
	return Version + "+" + BuildTime
}
