package version

import _ "embed"

//go:generate bash generate_version.sh
//go:embed version.txt
var version string

// Version returns the version string.
func Version() string {
	return version
}
