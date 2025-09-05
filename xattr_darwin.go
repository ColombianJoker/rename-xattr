//go:build darwin
// +build darwin

package main

import (
	"github.com/pkg/xattr"
)

// renameXattrOS provides a platform-specific implementation.
func renameXattrOS(path, oldName, newName string) error {
	xattrValue, err := xattr.Get(path, oldName)
	if err != nil {
		return err
	}
	err = xattr.Set(path, newName, xattrValue)
	if err != nil {
		return err
	}
	return xattr.Remove(path, oldName)
}
