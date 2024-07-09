package main

const (
	// Ext is the extension for EBO backup files
	Ext = ".xbk"
)

// IsFileXBK tests if file is a EBO Backup file by checking the file extension
func IsFileXBK(name string) bool {
	return hasExt(name, Ext)
}
