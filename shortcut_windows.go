//go:build windows
// +build windows

package xcpdir

import (
	"path/filepath"

	lnk "github.com/parsiya/golnk"
)

func shortcutEntity(shortcutPath string) (string, error) {
	// Lnk, err := lnk.File("test.lnk")
	Lnk, err := lnk.File(shortcutPath)
	if err != nil {
		return "", err
	}

	// fmt.Println(Lnk.Header)
	// fmt.Println(Lnk.LinkInfo)
	// fmt.Println(Lnk.StringData)
	// fmt.Println(Lnk.DataBlocks)

	// fmt.Println("BasePath", Lnk.LinkInfo.LocalBasePath)
	// fmt.Println("CommonPathSuffix", Lnk.LinkInfo.CommonPathSuffix)
	return filepath.Join(Lnk.LinkInfo.LocalBasePath, Lnk.LinkInfo.CommonPathSuffix), nil
}
