// Package fs provides some common utilities which GoLang developers use when working with files, either system files or web files
package fs

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"mime"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	// Version current version number
	Version = "0.0.5"
)

// PathSeparator is the OS-specific path separator
var PathSeparator = string(os.PathSeparator)

// DirectoryExists returns true if a directory(or file) exists, otherwise false
func DirectoryExists(dir string) bool {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return false
	}
	return true
}

// GetHomePath returns the user's $HOME directory
func GetHomePath() string {
	if runtime.GOOS == "windows" {
		return os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
	}
	return os.Getenv("HOME")
}

// GetParentDir returns the parent directory(string) of the passed targetDirectory (string)
func GetParentDir(targetDirectory string) string {
	lastSlashIndex := strings.LastIndexByte(targetDirectory, os.PathSeparator)
	//check if the slash is at the end , if yes then re- check without the last slash, we don't want /path/to/ , we want /path/to in order to get the /path/ which is the parent directory of the /path/to
	if lastSlashIndex == len(targetDirectory)-1 {
		lastSlashIndex = strings.LastIndexByte(targetDirectory[0:lastSlashIndex], os.PathSeparator)
	}

	parentDirectory := targetDirectory[0:lastSlashIndex]
	return parentDirectory
}

var (
	// errFileOpen describes the error when opening a file.
	errFileOpen = errors.New("open file")
	// errFileCreate describes the error when creating a file.
	errFileCreate = errors.New("create file")
	// errFileRemove describes the error when deleting a file.
	errFileRemove = errors.New("delete file")
	// errFileCopy decsribes the error when copying a file.
	errFileCopy = errors.New("copy file")
	// errDirCreate describes the error when creating a directory.
	errDirCreate = errors.New("create directory")
	// errNotDir describes the error when a source file is not a directory.
	errNotDir = errors.New("source is not a directory")
	// errFileRead describes the error when reading a file.
	errFileRead = errors.New("read file")
)

// RemoveFile removes a file or directory and returns an error, if any
func RemoveFile(filePath string) error {
	err := os.RemoveAll(filePath)
	if err != nil {
		return fmt.Errorf("%s: %w", filePath, errFileRemove)
	}
	return nil
}

// RenameDir renames (moves) oldpath to newpath.
// If newpath already exists, Rename replaces it.
// OS-specific restrictions may apply when oldpath and newpath are in different directories.
// If there is an error, it will be of type *LinkError.
//
// It's a copy of os.Rename
func RenameDir(oldPath string, newPath string) error {
	return os.Rename(oldPath, newPath)
}

// CopyFile accepts full path of the source and full path of destination, if file exists it's overrides it
// this function doesn't checks for permissions and all that, it returns an error
func CopyFile(source string, destination string) error {
	reader, err := os.Open(source)

	if err != nil {
		return fmt.Errorf("%w: %s", errFileCopy, err.Error())
	}

	defer reader.Close()

	writer, err := os.Create(destination)
	if err != nil {
		return fmt.Errorf("%w: %s", errFileCreate, err.Error())
	}

	defer writer.Close()

	_, err = io.Copy(writer, reader)
	if err != nil {
		return fmt.Errorf("%w: %s", errFileCopy, err.Error())
	}

	err = writer.Sync()
	if err != nil {
		return fmt.Errorf("%w: %s", errFileCopy, err.Error())
	}

	return nil
}

// CopyDir recursively copies a directory tree, attempting to preserve permissions.
// Source directory must exist.
func CopyDir(source string, dest string) (err error) {

	// get properties of source dir
	fi, err := os.Stat(source)
	if err != nil {
		return err
	}

	if !fi.IsDir() {
		return fmt.Errorf("%s: %w", source, errNotDir)
	}

	// create dest dir

	err = os.MkdirAll(dest, fi.Mode())
	if err != nil {
		return err
	}

	entries, err := os.ReadDir(source)

	for _, entry := range entries {

		sfp := source + PathSeparator + entry.Name()
		dfp := dest + PathSeparator + entry.Name()
		if entry.IsDir() {
			err = CopyDir(sfp, dfp)
			if err != nil {
				return
			}
		} else {
			// perform copy
			err = CopyFile(sfp, dfp)
			if err != nil {
				return
			}
		}

	}
	return
}

// Unzip extracts a zipped file to the target location
// returns the path of the created folder (if any) and an error (if any)
func Unzip(archive string, target string) (string, error) {
	reader, err := zip.OpenReader(archive)
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(target, 0755); err != nil {
		return "", fmt.Errorf("%w: %s", errDirCreate, err.Error())
	}
	createdFolder := ""
	for _, file := range reader.File {
		path := filepath.Join(target, file.Name)
		if file.FileInfo().IsDir() {
			os.MkdirAll(path, file.Mode())
			if createdFolder == "" {
				// this is the new directory that zip has
				createdFolder = path
			}
			continue
		}

		fileReader, err := file.Open()
		if err != nil {
			return "", fmt.Errorf("%w: %s", errFileOpen, err.Error())
		}
		defer fileReader.Close()

		targetFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			return "", fmt.Errorf("%w: %s", errFileOpen, err.Error())
		}
		defer targetFile.Close()

		if _, err := io.Copy(targetFile, fileReader); err != nil {
			return "", fmt.Errorf("%w: %s", errFileCopy, err.Error())
		}

	}

	reader.Close()
	return createdFolder, nil
}

// TypeByExtension returns the MIME type associated with the file extension ext.
// The extension ext should begin with a leading dot, as in ".html".
// When ext has no associated type, TypeByExtension returns "".
//
// Extensions are looked up first case-sensitively, then case-insensitively.
//
// The built-in table is small but on unix it is augmented by the local
// system's mime.types file(s) if available under one or more of these
// names:
//
//	/etc/mime.types
//	/etc/apache2/mime.types
//	/etc/apache/mime.types
//
// On Windows, MIME types are extracted from the registry.
//
// Text types have the charset parameter set to "utf-8" by default.
func TypeByExtension(fullfilename string) (t string) {
	ext := filepath.Ext(fullfilename)
	//these should be found by the windows(registry) and unix(apache) but on windows some machines have problems on this part.
	if t = mime.TypeByExtension(ext); t == "" {
		// no use of map here because we will have to lock/unlock it, by hand is better, no problem:
		if ext == ".json" {
			t = "application/json"
		} else if ext == ".js" {
			t = "application/javascript"
		} else if ext == ".zip" {
			t = "application/zip"
		} else if ext == ".3gp" {
			t = "video/3gpp"
		} else if ext == ".7z" {
			t = "application/x-7z-compressed"
		} else if ext == ".ace" {
			t = "application/x-ace-compressed"
		} else if ext == ".aac" {
			t = "audio/x-aac"
		} else if ext == ".ico" { // for any case
			t = "image/x-icon"
		} else if ext == ".png" {
			t = "image/png"
		} else {
			t = "application/octet-stream"
		}
		// mime.TypeByExtension returns as text/plain; | charset=utf-8 the static .js (not always)
	} else if t == "text/plain" || t == "text/plain; charset=utf-8" {
		if ext == ".js" {
			t = "application/javascript"
		}
	}
	return
}
