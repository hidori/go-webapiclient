package meta

import (
	_ "embed"
	"strings"
)

//go:embed version.txt
var version string

func GetVersion() string {
	return "v" + strings.TrimSpace(version)
}
