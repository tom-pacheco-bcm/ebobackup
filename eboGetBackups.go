/*
EBO Backup Tool

SE Building Operations Enterprise Server Backup Collection Utility.

Collects latest backups from Enterprise Server and copies to a local backup folder.
Makes a Zip Archive file with latest backups.
*/
package main

import (
	_ "embed"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

//go:embed version.txt
var version string
var exeDir = filepath.Dir(os.Args[0])
var exeName = filepath.Base(os.Args[0])
var configName = changeExt(exeName, ".config")
var defaultConfigFile = filepath.Join(exeDir, configName)

var logFile string

// visitLatestBackupFiles returns a WalkFunc to build a file list with the
// latest backup file from each directory
func visitLatestBackupFiles(files *[]string) filepath.WalkFunc {
	var currentDir string
	var latestTime time.Time

	return func(path string, info os.FileInfo, err error) error {

		if err != nil {
			log.Fatal(err)
		}

		if IsFileXBK(path) {

			dirName := filepath.Dir(path)
			modTime := info.ModTime()

			switch {
			case currentDir != dirName:
				currentDir = dirName
				latestTime = modTime
				*files = append(*files, path)

			case modTime.After(latestTime):
				latestTime = modTime
				(*files)[len(*files)-1] = path
			}
		}

		return nil
	}
}

var ErrMissingConfigFile = errors.New("file not found")

// getConfigFile returns the config file. if one is passed
// by argument then it is returned. otherwise the default
// config is returned
func getConfigFile() (string, error) {

	if configName == "" {
		return configName, ErrMissingConfigFile
	}

	if _, err := os.Stat(configName); err == nil {
		return configName, nil
	}

	if filepath.IsAbs(configName) {
		return configName, ErrMissingConfigFile
	}

	name := filepath.Join(exeDir, configName)
	if _, err := os.Stat(name); err != nil {
		return name, ErrMissingConfigFile
	}
	return name, nil
}

// StringPredicate is a predicate function for strings
type StringPredicate func(string) bool

// filter the list based on predicate function and return.
// modifies the list in place. returns the slice with the
// matching items from the start of the provided slice/array.
func filter(xs []string, predicate StringPredicate) []string {
	count := 0
	for i, x := range xs {
		if predicate(x) {
			xs[count], xs[i] = x, xs[count]
			count++
		}
	}
	return xs[:count]
}

func readDirNames(path string) []string {
	dir, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer dir.Close()
	names, err := dir.Readdirnames(-1)
	if err != nil {
		log.Fatal(err)
	}
	return names
}

func readDir(path string) []fs.FileInfo {
	dir, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer dir.Close()
	fis, err := dir.Readdir(-1)
	if err != nil {
		log.Fatal(err)
	}
	return fis
}

// fileExists tests if exists in list
func fileExists(n string, ns []string) bool {
	for _, f := range ns {
		if filepath.Base(n) == filepath.Base(f) {
			return true
		}
	}
	return false
}

func openLog() io.Closer {

	if logFile == "" {
		return nil
	}

	logOut, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE, fs.ModePerm)

	if err != nil {
		return nil
	}

	log.SetOutput(logOut)
	return logOut
}

var root = &cobra.Command{
	Use:   "ebobackup",
	Short: "a tool to collect ebo backups",
	Long: `Get latest EBO Backups and copy to a specified folder.

	searches for the default config file if it is not provided. 
	`,
	Run: func(cmd *cobra.Command, args []string) {

		logger := openLog()
		if logger != nil {
			defer logger.Close()
		}

		err := backupAndArchive()
		if err != nil {
			cmd.Usage()
		}
	},
}

var findCmd = &cobra.Command{
	Use:   "find",
	Short: "list the ebo servers backup locations",
	Run: func(cmd *cobra.Command, args []string) {
		listLocations()
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "print the version of the command",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("%s version %s", exeName, version)
	},
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "list the ebo backups",
	Run: func(cmd *cobra.Command, args []string) {
		file, err := getConfigFile()
		if err != nil {
			log.Printf("Error config file '%s' not found!\n", file)
			cmd.Usage()
			return
		}
		config := loadConfig(file)
		files := config.getBackupFiles()
		for _, file := range files {
			fmt.Println(file)
		}
	},
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "create an initial backup configuration file",
	Run: func(cmd *cobra.Command, args []string) {
		file, err := getConfigFile()
		if err != ErrMissingConfigFile {
			log.Printf("Error config file '%s' already exists!\n", file)
			cmd.Usage()
			return
		}
		initializeConfig()
	},
}

func main() {

	root.AddCommand(findCmd)
	root.AddCommand(versionCmd)
	root.AddCommand(listCmd)
	root.AddCommand(initCmd)

	root.Flags().StringVar(&logFile, "log", "", "optional log file")
	root.Flags().StringVar(&configName, "config", configName, "configuration file")

	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func backupAndArchive() error {

	log.Printf("starting backup\n")
	defer func() {
		log.Printf("backup completed\n")
	}()

	file, err := getConfigFile()
	if err != nil {
		log.Printf("Error config file '%s' not found!\n", file)
		return err
	}
	config := loadConfig(file)

	log.Printf("checking backups in %s\n", config.ESBackupPath)
	files := config.getBackupFiles()
	log.Printf("found %d backups\n", len(files))

	config.collectBackups(files)

	if !config.Archive {
		return nil
	}

	log.Printf("starting archive\n")
	fileName := config.archiveBackups(files)
	log.Printf("archive complete\n")

	if !config.Ftp {
		return nil
	}

	if !config.isFtpScheduled() {
		log.Printf("ftp not scheduled\n")
		return nil
	}

	log.Printf("uploading archive to ftp\n")
	config.uploadArchive(fileName)
	return nil
}

// getBackupFiles gets the latest set of backup files from the backup path
func (config *configSettings) getBackupFiles() []string {
	files := []string{}
	err := filepath.Walk(config.ESBackupPath, visitLatestBackupFiles(&files))
	if err != nil {
		log.Fatal(err)
	}
	return files
}

// collectBackups reads all sub folders
func (config *configSettings) collectBackups(files []string) {

	err := os.MkdirAll(filepath.FromSlash(config.BackupFolder), fs.ModePerm|fs.ModeDir)
	if err != nil {
		log.Fatal(err)
	}

	// delete old .xbk backup files.
	// keeping current files so we don't need to copy again
	// comparison is by name only
	names := readDirNames(config.BackupFolder)
	names = filter(names, IsFileXBK)
	for _, file := range names {
		if !fileExists(file, files) {
			log.Printf("deleting old backup `%s` from %s\n", file, config.BackupFolder)
			err := os.Remove(filepath.Join(config.BackupFolder, file))
			if err != nil {
				log.Printf("error deleting backup: %v", err)
			}
		}
	}

	for _, file := range files {
		if fileExists(file, names) {
			continue
		}
		log.Printf("copying `%s`\n", filepath.Base(file))
		newFile, err := copyFileTo(file, config.BackupFolder)
		if err != nil {
			return
		}

		validateFile(newFile)
	}
}

// archiveBackups creates a new archive file with the current backups
func (config *configSettings) archiveBackups(files []string) string {

	if config.ArchiveFolder == "" {
		log.Fatal("error, no archive folder.")
	}

	err := os.MkdirAll(config.ArchiveFolder, fs.ModePerm|fs.ModeDir)
	if err != nil {
		log.Fatal(err)
	}

	fileName := config.getZipFile()
	log.Printf("creating archive `%s`\n", fileName)
	ZipFiles(fileName, files)

	config.archiveRemoveOld()

	return fileName
}

// archiveRemoveOld removes old archives
func (config *configSettings) archiveRemoveOld() {

	log.Printf("removing old archives")

	if config.ArchiveFolder == "" {
		log.Fatal("error, no archive folder.")
	}
	if config.ArchiveCount < 1 {
		return
	}

	fis := readDir(config.ArchiveFolder)
	if len(fis) <= config.ArchiveCount {
		return
	}

	sort.Slice(fis, func(i, j int) bool { return fis[i].ModTime().After(fis[j].ModTime()) })

	fis = fis[config.ArchiveCount:]
	for _, fi := range fis {
		log.Printf("removing %q", fi.Name())
		err := os.Remove(filepath.Join(config.ArchiveFolder, fi.Name()))
		if err != nil {
			log.Printf("error [%v] removing archive `%s`\n", err, fi.Name())
		}
	}
}

// getZipFile generates a zip-file name from config and the current date
func (config *configSettings) getZipFile() string {

	currentTime := time.Now()

	zipFile := config.ArchiveName
	zipExt := filepath.Ext(zipFile)

	if zipExt == "" {
		zipExt = ".zip"
	} else {
		zipFile = strings.TrimSuffix(zipFile, zipExt)
	}

	if config.ArchiveISOWeek {
		isoYear, isoWeek := currentTime.ISOWeek()
		zipFile = fmt.Sprintf("%s_%04dW%02d%s", zipFile, isoYear, isoWeek, zipExt)
	} else {
		if config.ArchiveAddYear {
			currentYear := currentTime.Year()
			zipFile = fmt.Sprintf("%s_%04d", zipFile, currentYear)

		}
		if config.ArchiveAddMonth {
			currentMonth := currentTime.Month()
			zipFile = fmt.Sprintf("%s_%02d", zipFile, currentMonth)

		}
		zipFile = fmt.Sprintf("%s%s", zipFile, zipExt)
	}

	return filepath.Join(config.ArchiveFolder, zipFile)
}

// isFtpScheduled checks if ftp is scheduled
func (config *configSettings) isFtpScheduled() bool {

	if config.FtpWeekday == "" {
		return true
	}

	currentTime := time.Now()
	weekday := strings.ToLower(currentTime.Weekday().String())

	return config.FtpWeekday == weekday
}

func Usage() {

	fmt.Printf(`%[1]s
	Get latest EBO Backups and copy to a specified folder.

Usage:
	%[1]s [config]

searches for the default config file '%[2]s' if it is not provided. 

Sample Config: 
ESBackupPath    = "C:\ProgramData\Schneider Electric EcoStruxure\Building Operation 6.0\Enterprise Server\db_backup" 
BackupFolder    = "D:\backups\db_backup"
ArchiveFolder   = "D:\backups\archives" 
Archive         = True
ArchiveCount    = 5
ArchiveName     = "my_site_backups"
ArchiveISOWeek  = True
ArchiveAddYear  = False
ArchiveAddMonth = False

Examples:
	> %[1]s
	`, exeName, configName)
}
