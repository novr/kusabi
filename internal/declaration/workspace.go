package declaration

import (
	"github.com/novr/kusabi/internal/manifest"
)

// OpenWorkspace finds and opens kusabi.yaml from startDir or its parents.
func OpenWorkspace(startDir string) (*manifest.File, error) {
	path, err := manifest.Find(startDir)
	if err != nil {
		return nil, err
	}
	return manifest.Open(path)
}
