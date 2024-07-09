package main

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// copy a file to the destination directory
func copyFileTo(sourceFile, destDir string) (string, error) {

	// get new filename and check if doesn't exist
	destFile := filepath.Join(destDir, filepath.Base(sourceFile))
	if _, err := os.Stat(destFile); err == nil {
		return destFile, err // file exists no need to copy
	}

	err := copyFile(destFile, sourceFile)
	if err != nil {
		log.Fatal(err)
	}

	err = copyInfo(destFile, sourceFile)
	if err != nil {
		log.Fatal(err)
	}

	return destFile, err
}

func copyInfo(destFile, sourceFile string) error {

	info, err := os.Stat(sourceFile)
	if err != nil {
		return err
	}

	modeTime := info.ModTime()
	os.Chtimes(destFile, modeTime, modeTime)
	return nil
}

// copyFile copies a file to the destination file
func copyFile(dstName, srcName string) error {

	src, err := os.Open(srcName)
	if err != nil {
		return err
	}
	defer src.Close()

	// Create new file
	dst, err := os.Create(dstName)
	if err != nil {
		return err
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	if err != nil {
		return err
	}

	return nil
}

// remove the file extension from the path
func removeExt(path string) string {
	return strings.TrimSuffix(path, filepath.Ext(path))
}

// change the file extension from the path
func changeExt(path, newExt string) string {
	return removeExt(path) + newExt
}

func hasExt(path, ext string) bool {
	return strings.EqualFold(filepath.Ext(path), ext)
}
