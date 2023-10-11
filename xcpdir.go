package xcpdir

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func Args(
	src string, // source directory
	dst string, // destination directory
	trash string,
	dryrun bool,
	hidden bool,
	modified bool,
	link bool,
	ignore bool,
	backup bool,
	sync bool,
	// message chan string,
) *args {
	message := make(chan []string)
	return &args{
		src:      src,
		dst:      dst,
		trash:    trash,
		dryrun:   dryrun,
		hidden:   hidden,
		modified: modified,
		link:     link,
		ignore:   ignore,
		backup:   backup,
		sync:     sync,
		message:  message,
	}
}

type args struct {
	src      string // source directory
	dst      string // destination directory
	trash    string
	dryrun   bool
	hidden   bool
	modified bool
	link     bool
	ignore   bool
	backup   bool
	sync     bool
	message  chan []string
}

func (a *args) Message() chan []string {
	return a.message
}

func (a *args) sendMessage(o ...string) {
	// debug
	// _, file, line, _ := runtime.Caller(1)
	// o = append(o, fmt.Sprintf("(file: %s, line: %d)", file, line))
	a.message <- o
}

// Make directory and Copy a file
// Check dryrun flag
func (a *args) copyFile(srcName, dstName string) error {
	info, err := os.Stat(srcName)
	if err != nil {
		return newError(err)
	}

	// Make directory
	dir, _ := filepath.Split(dstName)
	if !fileExists(dir) {
		a.sendMessage("mkdir", dir)
		if !a.dryrun {
			err := os.MkdirAll(dir, os.ModePerm)
			if err != nil {
				return newError(err)
			}
		}
	}

	a.sendMessage("copy", srcName, dstName)
	if a.dryrun {
		return nil
	}

	src, err := os.Open(srcName)
	if err != nil {
		return newError(err)
	}
	defer src.Close()

	dst, err := os.Create(dstName)
	if err != nil {
		return newError(err)
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	if err != nil {
		return newError(err)
	}

	err = os.Chtimes(dstName, info.ModTime(), info.ModTime())
	if err != nil {
		a.sendMessage(err.Error())
	}
	return nil
}

type opMode string

const (
	modeBackup opMode = "backup"
	modeSync   opMode = "sync"
)

// Backup a file
// Copy a file
// Remove a source file if sync mode
func (a *args) backupFile(mode opMode, path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return newError(err)
	}

	dir, base := filepath.Split(path)
	// fname, ext := splitBase(base)
	// backup := filepath.Join(*a.trash, dir[len(*a.dst):], fname+info.ModTime().Format("_2006-01-02_150405")+ext)
	backup := filepath.Join(a.trash, dir[len(a.dst):], backupBaseName(base, info.ModTime()))
	a.sendMessage(string(mode), path, backup)
	err = a.copyFile(path, backup)
	if err != nil {
		return newError(err)
	}

	// Sync remove dst file
	if a.sync && mode == modeSync {
		a.sendMessage("remove", path)
		if !a.dryrun {
			err := os.Remove(path)
			if err != nil {
				return newError(err)
			}
		}
	}
	return newError(err)
}

// - Search for symbolic link entity
// - Backup,
// - Check exists file size modtime
// - Skip or Copy
func (a *args) copy(path string) error {
	// Create dst path
	dst := filepath.Join(a.dst, path[len(a.src):])
	a.sendMessage("target", path, dst)

	// Symbolic link
	entity, err := filepath.EvalSymlinks(path)
	if err != nil {
		return newError(err)
	}

	// Backup
	if a.backup && fileExists(dst) {
		info1, err1 := os.Stat(entity)
		if err1 != nil {
			return newError(err1)
		}
		info2, err2 := os.Stat(dst)
		if err2 != nil {
			return newError(err2)
		}
		if a.modified && info1.Size() == info2.Size() && info1.ModTime().Compare(info2.ModTime()) == 0 {
			a.sendMessage("skip", entity, dst)
			return nil
		}
		err = a.backupFile(modeBackup, dst)
		if err != nil {
			return newError(err)
		}
	}

	return a.copyFile(entity, dst)
}

// OS Windows only
func (a *args) windowsShortcut(path string) error {
	ext := filepath.Ext(path)
	if strings.ToLower(ext) != ".lnk" {
		return nil
	}
	entity, err := shortcutEntity(path)
	if err != nil {
		return newError(err)
	}
	info, err := os.Stat(entity)
	if err != nil {
		return newError(err)
	}

	// Create dst path
	// dst := filepath.Join(*a.dst, path[len(*a.src):])
	//fmt.Printf("entity %s", entity)
	dst := filepath.Join(a.dst, filepath.Dir(path[len(a.src):]), filepath.Base(entity))
	//fmt.Printf("dst %s", dst)
	a.sendMessage("wlnk", path, entity)

	if info.IsDir() {
		var a2 args = *a
		a2.src = entity
		a2.dst = dst
		err = a2.syncDir(a2.src)
		if err != nil {
			return newError(err)
		}
		err = a2.dirWalk(a2.src)
		if err != nil {
			return newError(err)
		}
	} else {
		// Backup
		if a.backup && fileExists(dst) {
			info1, err1 := os.Stat(entity)
			if err1 != nil {
				return newError(err1)
			}
			info2, err2 := os.Stat(dst)
			if err2 != nil {
				return newError(err2)
			}
			if a.modified && info1.Size() == info2.Size() && info1.ModTime().Compare(info2.ModTime()) == 0 {
				a.sendMessage("skip", entity, dst)
				return nil
			}
			a.backupFile(modeBackup, dst)
		}
		a.copyFile(entity, dst)
	}
	return nil
}

// Check sync flag and sync directory
func (a *args) syncDir(srcDir string) error {
	if !a.sync {
		return nil
	}

	dstDir := filepath.Join(a.dst, srcDir[len(a.src):])
	if !fileExists(dstDir) {
		return nil
	}
	files, err := os.ReadDir(dstDir)
	if err != nil {
		if a.ignore {
			a.sendMessage(err.Error())
			return nil
		}
		return newError(err)
	}
	a.sendMessage("sync", srcDir, dstDir)
	for _, file := range files {
		srcPath := filepath.Join(srcDir, file.Name())
		if fileExists(srcPath) {
			continue
		}

		info, err := os.Stat(dstDir)
		if err != nil {
			if a.ignore {
				a.sendMessage(err.Error())
				continue
			}
			return newError(err)
		}
		if info.IsDir() {
			// no backup, no sync
			// backup, copy and remove
			var a2 args = *a
			a2.backup = false
			a2.sync = false
			a2.link = false
			// *a2.src = filepath.Join(srcDir, file.Name())
			a2.src = filepath.Join(dstDir, file.Name())
			a2.dst = filepath.Join(a.trash, dstDir[len(a.dst):], backupBaseName(file.Name(), info.ModTime()))
			// fmt.Println("******", *a2.src, *a2.dst)
			a.sendMessage(string(modeBackup), a2.src, a2.dst)
			err := a2.dirWalk(a2.src)
			if err != nil {
				if a2.ignore {
					a2.sendMessage(err.Error())
					continue
				}
				return newError(err)
			}
			a.sendMessage("remove", a2.src)
			if !a.dryrun {
				err := os.RemoveAll(a2.src)
				if err != nil {
					if a2.ignore {
						a2.sendMessage(err.Error())
						continue
					}
					return newError(err)
				}
			}
		} else {
			dstPath := filepath.Join(dstDir, file.Name())
			err := a.backupFile(modeSync, dstPath) // copy and remove
			if err != nil {
				if a.ignore {
					a.sendMessage(err.Error())
					continue
				}
				return newError(err)
			}
		}
	} // for

	if err != nil && a.ignore {
		a.sendMessage(err.Error())
		return nil
	}
	return newError(err)
}

func (a *args) dirWalk(path string) error {
	// _, file, line, _ := runtime.Caller(1)
	// fmt.Printf("(file: %s, line: %d)\n", file, line)
	files, err := os.ReadDir(path)
	if err != nil {
		if a.ignore {
			a.sendMessage(err.Error())
			return nil
		}
		return newError(err)
	}

	for _, file := range files {
		path := filepath.Join(path, file.Name())

		info, err := os.Stat(path)
		if err != nil {
			if a.ignore {
				a.sendMessage(err.Error())
				continue
			}
			return newError(err)
		}

		hidden, err := isHidden(path)
		if err != nil {
			if a.ignore {
				a.sendMessage(err.Error())
				continue
			}
			return newError(err)
		}
		if !a.hidden && hidden {
			continue
		}

		if info.IsDir() {
			err := a.syncDir(path)
			if err != nil {
				if a.ignore {
					a.sendMessage(err.Error())
					continue
				}
				return newError(err)
			}
			if err := a.dirWalk(path); err != nil {
				if a.ignore {
					a.sendMessage(err.Error())
					continue
				}
				return newError(err)
			}
		} else {
			ext := filepath.Ext(path)
			if strings.ToLower(ext) == ".lnk" && runtime.GOOS == "windows" {
				err = a.windowsShortcut(path)
			} else {
				err = a.copy(path)
			}
			if err != nil {
				if a.ignore {
					a.sendMessage(err.Error())
					continue
				}
				return newError(err)
			}
		}
	} // for

	if err != nil && a.ignore {
		a.sendMessage(err.Error())
		return nil
	}
	return newError(err)
}

func (a *args) Xcpdir() error {
	info, err := os.Stat(a.src)
	if err != nil {
		return newError(err)
	}
	if !info.IsDir() {
		return newError(fmt.Errorf("%s is not a directory", a.src))
	}
	err = a.syncDir(a.src)
	if err != nil {
		return newError(err)
	}
	return a.dirWalk(a.src)
}
