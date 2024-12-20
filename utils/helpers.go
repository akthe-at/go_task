package utils

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

// CheckIfProjDir checks if the current directory is a project directory
// by checking for the presence of a .git directory in
// the current directory or any of its parent directories.
func CheckIfProjDir() (bool, string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return false, "", fmt.Errorf("fn - checkIfProjDir: error getting current directory: (%s): %v", dir, err)
	}

	for {
		files, err := os.ReadDir(dir)
		if err != nil {
			return false, "", fmt.Errorf("fn - checkIfProjDir: error reading directory (%s) : %v", dir, err)
		}
		for _, file := range files {
			if file.Name() == ".git" {
				return true, dir, nil
			}
		}

		parentDir := filepath.Dir(dir)
		if parentDir == dir {
			break
		}
		dir = parentDir
	}

	return false, "", nil
}

// ExpandPath the path to include the user's home directory
func ExpandPath(path string) (string, error) {
	if strings.HasPrefix(path, "~") {
		currentUser, err := user.Current()
		if err != nil {
			return "", err
		}
		// path[1:] to remove the ~
		path = filepath.Join(currentUser.HomeDir, path[1:])
	}
	return path, nil
}
