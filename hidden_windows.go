//go:build windows
// +build windows

/*
This code for detecting hidden files on Windows is based on a code snippet
published under the MIT License at the following website:

https://freshman.tech/snippets/go/detect-hidden-file/
© 2016-2023 Ayooluwa Isaiah. All rights reserved.
Code snippets may be used freely under the terms of the MIT License.
*/

// hidden_windows.go

package xcpdir

import (
	"path/filepath"
	"syscall"
)

const dotCharacter = 46

// isHidden checks if a file is hidden on Windows.
func isHidden(path string) (bool, error) {
	// dotfiles also count as hidden (if you want)
	if path[0] == dotCharacter {
		return true, nil
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return false, err
	}

	// Appending `\\?\` to the absolute path helps with
	// preventing 'Path Not Specified Error' when accessing
	// long paths and filenames
	// https://docs.microsoft.com/en-us/windows/win32/fileio/maximum-file-path-limitation?tabs=cmd
	pointer, err := syscall.UTF16PtrFromString(`\\?\` + absPath)
	if err != nil {
		return false, err
	}

	attributes, err := syscall.GetFileAttributes(pointer)
	if err != nil {
		return false, err
	}

	return attributes&syscall.FILE_ATTRIBUTE_HIDDEN != 0, nil
}
