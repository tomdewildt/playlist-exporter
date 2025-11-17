//go:build windows

package io

import (
	"os"
	"runtime"
	"syscall"
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
	if runtime.GOOS == "windows" {
		if pointer, ok := info.Sys().(*syscall.Win32FileAttributeData); ok {
			return pointer.FileAttributes&syscall.FILE_ATTRIBUTE_HIDDEN != 0
		}
	}
	return false
}
