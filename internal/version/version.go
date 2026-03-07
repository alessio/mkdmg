package version

import (
	_ "embed"
	"strings"
)

//go:generate bash generate_version.sh
//go:embed version.txt
var version string

// Version returns the version string.
func Version() string {
	return strings.TrimSpace(version)
}
