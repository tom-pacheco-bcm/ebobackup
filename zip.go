package main

import (
	"archive/zip"
	"io"
	"log"
	"os"
	"path/filepath"
)

/*
ZipFiles compresses one or many files into a single zip archive file.

Param 1: filename is the output zip file's name.
Param 2: files is a list of files to add to the zip.
*/
func ZipFiles(filename string, files []string) error {

	newZipFile, err := os.Create(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer newZipFile.Close()

	zipWriter := zip.NewWriter(newZipFile)
	defer zipWriter.Close()

	// Add files to zip
	for _, file := range files {
		log.Printf("adding `%s` to archive\n", filepath.Base(file))
		if err = addFileToZip(zipWriter, file); err != nil {
			log.Fatal(err)
		}
	}
	return nil
}

func addFileToZip(zipWriter *zip.Writer, filename string) error {

	fileToZip, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer fileToZip.Close()

	// Get the file information
	info, err := fileToZip.Stat()
	if err != nil {
		log.Fatal(err)
	}

	header, err := zip.FileInfoHeader(info)
	if err != nil {
		log.Fatal(err)
	}

	header.Method = zip.Deflate

	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		log.Fatal(err)
	}

	_, err = io.Copy(writer, fileToZip)
	return err
}

func CheckZipFile(filename string) error {

	info, err := os.Stat(filename)
	if err != nil {
		log.Fatal(err)
	}

	zipFile, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer zipFile.Close()

	_, err = zip.NewReader(zipFile, info.Size())
	if err != nil {
		log.Fatal(err)
	}

	// for _, f := range zipReader.File {
	// 	f.Name
	// 	zipReader.Open()
	// }

	return nil
}
