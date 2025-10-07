//go:build windows
// +build windows

package web

import (
	"fmt"
)

// Start returns an error on Windows as web mode is not supported
func Start(port int) error {
	return fmt.Errorf("web mode is not supported on Windows")
}

// StartWithLogging returns an error on Windows as web mode is not supported
func StartWithLogging(port int, enableLogs bool) error {
	return fmt.Errorf("web mode is not supported on Windows")
}
