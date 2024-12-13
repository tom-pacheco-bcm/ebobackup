package main

import (
	"log"
	"os"
	"time"

	"github.com/secsy/goftp"
)

func (config *configSettings) uploadArchive(fileName string) {

	ftpConfig := goftp.Config{
		User:               config.FtpUser,
		Password:           config.FtpPass,
		ConnectionsPerHost: 10,
		Timeout:            10 * time.Second,
		// Logger:             os.Stderr,
	}
	client, err := goftp.DialConfig(ftpConfig, config.FtpUri)
	if err != nil {
		panic(err)
	}
	defer client.Close()

	// open source file
	srcFile, err := os.Open(fileName)
	if err != nil {
		log.Fatal(err)
	}
	defer srcFile.Close()

	destName := config.getFtpFile(fileName)
	log.Printf("uploading `%s`\n", destName)

	// create destination file
	err = client.Store(destName, srcFile)
	if err != nil {
		log.Fatal(err)
	}
}
