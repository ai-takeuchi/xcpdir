# xcpdir Package

- 2023-10-03 Under operation test

## Introduction

The `xcpdir` package is a Go library that provides functionality for copying and synchronizing directories with various options. It can be used to perform operations such as copying files, creating backups, and managing symbolic links within source and destination directories.

## Installation

You can install the `xcpdir` package by using Go modules. Add the following import statement to your Go code to use this package:

```go
import "github.com/your-import-path/xcpdir"
```

## Usage

### Creating an Arguments Struct

To use the `xcpdir` package, you first need to create an arguments struct by calling the `Args` function. This struct holds various options and settings for the copying and synchronization process.

```go
args := xcpdir.Args(
    src,      // source directory
    dst,      // destination directory
    trash,    // trash directory
    dryrun,   // dry run mode
    hidden,   // include hidden files
    modified, // only copy modified files
    link,     // manage symbolic links
    ignore,   // ignore errors
	backup:   // copy the modified files to the trash path before copying
	sync:     // if the source does not exist, move the destination file to the trash path
)
```

## Copying Files

The `xcpdir` package provides a `copy` function for copying files from the source directory to the destination directory. It handles the creation of directories, backup of existing files, and copying of new files.

```go
err := args.copy(srcName, dstName)
```

## Synchronizing Directories

You can use the `sync` function to synchronize two directories by making the contents of the destination directory match the source directory. It handles file and directory creation, backups, and removal of obsolete files.

```go
err := args.sync(srcDir)
```

## Windows Shortcut Handling (Windows Only)

For Windows users, the `windowsShortcut` function can be used to handle Windows shortcuts (`.lnk` files). It converts shortcuts to their target files or directories and copies them to the destination.

```go
err := args.windowsShortcut(path)
```

## Executing the Copying and Synchronization Process

Finally, use the `Xcpdir` function to start the copying and synchronization process. This function works on the source directory and recursively processes its contents.

```go
err := args.Xcpdir()
```

## Options

The `xcpdir` package supports various options for controlling the behavior of the copy and sync operations. You can configure these options when creating the arguments struct:

- `src` - Source directory.
- `dst` - Destination directory.
- `trash` - Trash directory for backups.
- `dryrun` - Dry run mode (simulates without actual file operations).
- `hidden` - Include hidden files.
- `modified` - Copy only modified files.
- `link` - Manage symbolic links.
- `ignore` - Ignore errors during the process.
- `backup` - Copy the modified files to the trash path before copying
- `sync` - If the source does not exist, move the destination file to the trash path

## Examples

For examples and usage:

```go
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"localhost/xcpdir"
)

func init() {
	flag.Usage = func() {
		base := filepath.Base(os.Args[0])
		fmt.Printf("Usage: %s <options>\n", base)
		flag.PrintDefaults()
		fmt.Println("")
		fmt.Println("e.g.")
		fmt.Printf("  %s -src=dir1 -dst=dir2 -trash=dir3 -dryrun=false\n", base)
		fmt.Println("")
	}
}

func main() {
	f, err := os.OpenFile("xcpdir.log", os.O_CREATE|os.O_APPEND|os.O_RDWR, os.ModePerm)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer f.Close()
	opts := &slog.HandlerOptions{
		AddSource: true,
	}
	logger := slog.New(slog.NewTextHandler(io.Writer(f), opts))
	logger.Info("--------")

	// flags
	aSrc := flag.String("src", "", "source directory")
	aDst := flag.String("dst", "", "destination directory")
	aTrash := flag.String("trash", "", "trash directory")
	aDryrun := flag.Bool("dryrun", true, "")
	aHidden := flag.Bool("hidden", true, "including hidden files")
	aModified := flag.Bool("modified", true, "modified files only")
	aLink := flag.Bool("link", true, "including link files")
	aIgnore := flag.Bool("ignore", true, "ignore error")
	aBackup := flag.Bool("backup", true, "copy the modified files to the trash path before copying")
	aSync := flag.Bool("sync", true, "if the source does not exist, move the destination file to the trash path")

	flag.Parse()

	var err2 error
	if *aSrc == "" {
		err = errors.Join(err, errors.New("command line argument src not specified"))
	}
	if *aDst == "" {
		err = errors.Join(err, errors.New("command line argument dst not specified"))
	}

	if *aSrc, err2 = filepath.Abs(*aSrc); err != nil {
		err = errors.Join(err, err2)
	}
	if *aDst, err2 = filepath.Abs(*aDst); err != nil {
		err = errors.Join(err, err2)
	}

	if *aTrash == "" {
		dir, _ := filepath.Split(*aDst)
		*aTrash = filepath.Join(dir, "#trash")
	}

	if err != nil {
		fmt.Println(err)
		fmt.Println("")
		flag.Usage()
		os.Exit(1)
	}

	a := xcpdir.Args(
		*aSrc,
		*aDst,
		*aTrash,
		*aDryrun,
		*aHidden,
		*aModified,
		*aLink,
		*aIgnore,
		*aBackup,
		*aSync,
	)

	// receive result chan
	go func() {
		for {
			select {
			case result := <-a.Message():
				var s string
				switch result[0] {
				case "target":
					s = fmt.Sprintf("===> %s => %s", result[1], result[2])
				case "copy":
					s = fmt.Sprintf("copy %s => %s", result[1], result[2])
				case "mkdir":
					s = fmt.Sprintf("mkdr %s", result[1])
				case "backup":
					s = fmt.Sprintf("bkup %s => %s", result[1], result[2])
				case "sync":
					s = fmt.Sprintf("sync %s => %s", result[1], result[2])
				case "skip":
					s = fmt.Sprintf("skip %s => %s", result[1], result[2])
				case "remove":
					s = fmt.Sprintf("dele %s", result[1])
				case "wlnk":
					s = fmt.Sprintf("wlnk %s => %s", result[1], result[2])
				default:
					s = strings.Join(result, ", ")
				}
				fmt.Println(s)
				logger.Info(s)
			default:
			}
		}
	}()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		err = a.Xcpdir()
		if err != nil {
			fmt.Println(err)
			logger.Error(err.Error())
			os.Exit(1)
		}
		wg.Done()
	}()

	wg.Wait()
	os.Exit(0)
}
```
