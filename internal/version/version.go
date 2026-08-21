// Package version owns the Silver interpreter's semantic version.
package version

import "fmt"

const (
	Major int64 = 0
	Minor int64 = 4
	Patch int64 = 0
)

// String returns the current version in MAJOR.MINOR.PATCH form.
func String() string {
	return fmt.Sprintf("%d.%d.%d", Major, Minor, Patch)
}
