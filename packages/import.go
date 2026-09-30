package packages

import (
	"fmt"
	"strings"
)

// ParseImport validates <package>:<module>[/<namespace>] without normalizing
// away invalid components. Public module names are case-sensitive and omit .slv.
func ParseImport(request string) (packageName, moduleName string, err error) {
	packageName, moduleName, found := strings.Cut(request, ":")
	if !found || !validPackageName(packageName) || moduleName == "" ||
		strings.ContainsAny(moduleName, ":\\\x00") || strings.HasSuffix(strings.ToLower(moduleName), ".slv") {
		return "", "", fmt.Errorf("invalid package import %q: expected <package>:<module>[/<namespace>] without .slv", request)
	}
	for _, component := range strings.Split(moduleName, "/") {
		if component == "" || component == "." || component == ".." || strings.TrimSpace(component) != component {
			return "", "", fmt.Errorf("invalid package import %q: module path must contain nonempty names without . or .. components", request)
		}
	}
	return packageName, moduleName, nil
}
