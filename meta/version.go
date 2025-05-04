// Package meta provides metadata information about the package.
package meta

import (
	_ "embed"
	"strings"
)

//go:embed version.txt
var version string

// GetVersion returns the version string with 'v' prefix.
func GetVersion() string {
	return "v" + strings.TrimSpace(version)
}
