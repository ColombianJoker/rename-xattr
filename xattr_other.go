//go:build !darwin && !linux

package main

import (
	"fmt"
)

// renameXattrOS provides a fallback for unsupported operating systems.
func renameXattrOS(path, oldName, newName string) error {
	return fmt.Errorf("extended attribute renaming is not supported on this operating system")
}
