package main

import (
	"bytes"
	"log"
	"os"
	"strings"
)

type configSettings struct {
	ESBackupPath    string
	BackupFolder    string
	ArchiveFolder   string
	ArchiveName     string
	Archive         bool
	ArchiveAddYear  bool
	ArchiveAddMonth bool
	ArchiveISOWeek  bool
	Ftp             bool
	FtpUri          string
	FtpUser         string
	FtpPass         string
	FtpWeekday      string
}

func parseConfig(file []byte) configSettings {

	settings := map[string]string{}
	lines := bytes.Split(file, []byte{'\n'})

	for _, line := range lines {
		kv := bytes.Split(line, []byte{'='})
		if len(kv) == 2 {
			val := string(bytes.Trim(bytes.Split(kv[1], []byte{'#'})[0], "\" \n\r"))
			settings[string(bytes.ToLower(bytes.TrimSpace(kv[0])))] = val
		}
	}

	return configSettings{
		ESBackupPath: settings["esbackuppath"],
		BackupFolder: settings["backupfolder"],

		Archive:         strings.ToLower(settings["archive"]) == "true",
		ArchiveFolder:   settings["archivefolder"],
		ArchiveName:     settings["archivename"],
		ArchiveAddYear:  strings.ToLower(settings["archiveaddyear"]) == "true",
		ArchiveAddMonth: strings.ToLower(settings["archiveaddmonth"]) == "true",
		ArchiveISOWeek:  strings.ToLower(settings["archiveisoweek"]) == "true",

		Ftp:        strings.ToLower(settings["ftp"]) == "true",
		FtpUri:     settings["ftpuri"],
		FtpUser:    settings["ftpuser"],
		FtpPass:    settings["ftppass"],
		FtpWeekday: settings["ftpweekday"],
	}
}

func loadConfig(configFile string) configSettings {

	file, e := os.ReadFile(configFile)
	if e != nil {
		log.Fatal(e)
	}
	return parseConfig(file)

}
