package filepath

import (
	"os"
	"path/filepath"
)

// getSubFolders recursively retrieves all subfolders of the specified path.
func GetSubFolders(path string) (paths []string, err error) {
	err = filepath.Walk(path, func(newPath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			paths = append(paths, newPath)
		}
		return nil
	})
	return paths, err
}
