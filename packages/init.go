package packages

import (
	"fmt"
	"os"
	"path/filepath"
)

const manifestFilename = "package.yaml"

// Init creates an empty package manifest in directory.
func Init(directory, packageName string) (string, error) {
	if !validPackageName(packageName) {
		return "", fmt.Errorf("package name %q is invalid", packageName)
	}

	path := filepath.Join(directory, manifestFilename)
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		if os.IsExist(err) {
			return "", fmt.Errorf("%s already exists", path)
		}
		return "", fmt.Errorf("could not create %s: %w", path, err)
	}

	contents := fmt.Sprintf("package: %s\nexport: []\n", packageName)
	if _, err := file.WriteString(contents); err != nil {
		file.Close()
		os.Remove(path)
		return "", fmt.Errorf("could not write %s: %w", path, err)
	}
	if err := file.Close(); err != nil {
		os.Remove(path)
		return "", fmt.Errorf("could not write %s: %w", path, err)
	}

	return path, nil
}
