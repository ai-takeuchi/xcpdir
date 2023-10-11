package xcpdir

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"syscall"
	"time"
)

func newError(err error) error {
	if err != nil {
		_, file, line, _ := runtime.Caller(1)
		return fmt.Errorf("file: %s, line: %d, %s", file, line, err)
	}
	return nil
}

// Create backup name with base name
func backupBaseName(base string, modTime time.Time) string {
	fname, ext := splitBase(base)
	return fname + modTime.Format("_2006-01-02_150405") + ext
}

/*
func backupName(path string) (string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return "", err
	}
	dir, base := filepath.Split(path)
	fname, ext := splitBase(base)
	return filepath.Join(dir, fname+info.ModTime().Format("_2006-01-02_150405")+ext), nil
}
*/

// https://qiita.com/suin/items/b9c0f92851454dc6d461
func fileExists(filename string) bool {
	_, err := os.Stat(filename)

	if pathError, ok := err.(*os.PathError); ok {
		if pathError.Err == syscall.ENOTDIR {
			return false
		}
	}

	if os.IsNotExist(err) {
		return false
	}

	return true
}

// Split base name to name and ext
// e.g.
// - "example.ext"  => "example",".ext"
// - ".example.ext" => ".example",".ext"
// - ".example"     => ".example",""
func splitBase(base string) (string, string) {
	ext := filepath.Ext(base)
	if ext == base {
		return ext, ""
	}

	/*
		if strings.Index(base, ".") == 0 {
			return base, ""
		}
	*/

	// ext := filepath.Ext(base)
	return base[:len(base)-len(ext)], ext
}
