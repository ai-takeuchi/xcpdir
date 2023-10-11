//go:build !windows
// +build !windows

package xcpdir

import "errors"

func shortcutEntity(shortcutPath string) (string, error) {
	return "", errors.New("OS is not Windows")
}
