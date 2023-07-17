package main

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// re-path a file to a new directory
func changeDir(path, newDir string) string {
	return filepath.Join(newDir, filepath.Base(path))
}

// copy a file to the destination directory
func copyFileTo(sourceFile, destDir string) {

	// get new filename and check if doesn't exist
	destFile := changeDir(sourceFile, destDir)
	if _, err := os.Stat(destFile); err == nil {
		return // file exists no need to copy
	}

	original, err := os.Open(sourceFile)
	if err != nil {
		log.Fatal(err)
	}
	defer original.Close()

	// Create new file
	newFile, err := os.Create(destFile)
	if err != nil {
		log.Fatal(err)
	}
	defer newFile.Close()

	_, err = io.Copy(newFile, original)
	if err != nil {
		log.Fatal(err)
	}

	sourceInfo, err := os.Stat(sourceFile)
	if err != nil {
		log.Fatal(err)
	}
	modeTime := sourceInfo.ModTime()
	os.Chtimes(destFile, modeTime, modeTime)
}

// remove the file extension from the path
func removeExt(path string) string {
	return strings.TrimSuffix(path, filepath.Ext(path))
}

// change the file extension from the path
func changeExt(path, newExt string) string {
	return removeExt(path) + newExt
}

func isFileExt(path, ext string) bool {
	return strings.EqualFold(filepath.Ext(path), ext)
}
