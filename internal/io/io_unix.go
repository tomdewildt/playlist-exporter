//go:build aix || darwin || dragonfly || freebsd || js || linux || netbsd || openbsd || solaris

package io

import (
	"os"
	"path/filepath"
	"strings"
)

func CopyFile(source, destination string) error {
	input, err := os.ReadFile(source)
	if err != nil {
		return err
	}
	if err := os.WriteFile(destination, input, 0644); err != nil {
		return err
	}
	return nil
}

func EnsureDirExists(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return os.MkdirAll(path, 0755)
	}
	return nil
}

func IsHidden(info os.FileInfo, path string) bool {
	if strings.HasPrefix(filepath.Base(path), ".") {
		return true
	}
	return false
}
