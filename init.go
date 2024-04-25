package main

import (
	"fmt"
	"os"
)

func initializeConfig() {

	var c configSettings
	n := defaultConfigFile

	if _, err := os.Stat(n); err == nil {
		// config file exists return
		fmt.Println("exiting, config file already exists.")
		return
	}

	// get last ES backup location and use as the default
	ess, err := EnterpriseServers()
	if err != nil {
		return
	}
	if len(ess) > 0 {
		es := ess[len(ess)-1]
		dbPath, _ := es.DBBackupPath()
		c.ESBackupPath = dbPath
	}

	// set some default values
	c.BackupFolder = "c:\\ebobackup\\eb_backup"
	c.ArchiveFolder = "c:\\ebobackup\\archives"
	c.Archive = true
	c.ArchiveCount = 5
	c.ArchiveName = "my_site_backups"
	c.ArchiveISOWeek = true

	f, err := os.Create(n)
	if err != nil {
		return
	}

	_, _ = fmt.Fprintf(f, "ESBackupPath      = %q  # Enter the path to the ES backups 'db_backup'\n", c.ESBackupPath)
	_, _ = fmt.Fprintf(f, "BackupFolder      = %q  # the path to copy the backups to\n", c.BackupFolder)
	_, _ = fmt.Fprintf(f, "Archive           = %v  # flag to create archive file\n", c.Archive)
	_, _ = fmt.Fprintf(f, "ArchiveCount      = %v  # number of archive files to keep\n", c.ArchiveCount)
	_, _ = fmt.Fprintf(f, "ArchiveFolder     = %q  # the path to save an archive zip of all the backups\n", c.ArchiveFolder)
	_, _ = fmt.Fprintf(f, "ArchiveName       = %q  # name of the archive file\n", c.ArchiveName)
	_, _ = fmt.Fprintf(f, "ArchiveISOWeek    = %v  # add ISO week number to the archive name\n", c.ArchiveISOWeek)
	_, _ = fmt.Fprintf(f, "ArchiveAddYear    = %v  # add year to the archive name\n", c.ArchiveAddYear)
	_, _ = fmt.Fprintf(f, "ArchiveAddMonth   = %v  # add month to the archive name\n", c.ArchiveAddMonth)
	_, _ = fmt.Fprintf(f, "Ftp               = %v  # flag to upload to an ftp server\n", c.Ftp)
	_, _ = fmt.Fprintf(f, "FtpUri            = %q  # URI of the ftp server\n", c.FtpUri)
	_, _ = fmt.Fprintf(f, "FtpUser           = %q  # ftp user name\n", c.FtpUser)
	_, _ = fmt.Fprintf(f, "FtpPass           = %q  # ftp password\n", c.FtpPass)
	_, _ = fmt.Fprintf(f, "FtpWeekday        = %q  # day to upload the file\n", c.FtpWeekday)
	_, _ = fmt.Fprintf(f, "# Valid Weekdays  = Sunday, Monday, Tuesday, Wednesday, Thursday, Friday, Saturday\n")

}
