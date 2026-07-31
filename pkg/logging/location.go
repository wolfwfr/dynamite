package logging

import (
	"errors"
	"fmt"
	"os"
	"runtime"
)

// keeping around in case I change my mind, but /var/log is read-only for most
// users apparently.
func UserLogDir() (string, error) {
	var dir string

	switch runtime.GOOS {
	case "windows":
		dir = os.Getenv("AppData")
		if dir == "" {
			return "", errors.New("%AppData% is not defined")
		}

	default: // Unix & mac-os
		dir = "/var/log"
		info, err := os.Stat(dir)
		if err != nil {
			return "", fmt.Errorf("verifying '%s' exists %w", dir, err)
		}
		if !info.IsDir() {
			return "", fmt.Errorf("%s is not a directory", dir)
		}
	}

	return dir, nil
}
