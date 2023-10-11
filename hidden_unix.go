//go:build !windows
// +build !windows

package xcpdir

/*
This code for detecting hidden files on Windows is based on a code snippet
published under the MIT License at the following website:

https://freshman.tech/snippets/go/detect-hidden-file/
© 2016-2023 Ayooluwa Isaiah. All rights reserved.
Code snippets may be used freely under the terms of the MIT License.
*/

// hidden_unix.go

const dotCharacter = 46

func isHidden(path string) (bool, error) {
	if path[0] == dotCharacter {
		return true, nil
	}

	return false, nil
}
