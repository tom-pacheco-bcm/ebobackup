package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/sys/windows/registry"
)

// Computer\HKEY_LOCAL_MACHINE\SYSTEM\CurrentControlSet\Services\Building Operation 2.0 Enterprise Server
// ImagePath = "C:\Program Files (x86)\Schneider Electric EcoStruxure\Building Operation 2.0\Enterprise Server\bin\SE.SBO.EnterpriseServer.exe"

// C:\Program Files (x86)\Schneider Electric EcoStruxure\Building Operation 2.0\Enterprise Server\etc\dbpath.properties
// server.paths.db=C:/ProgramData/Schneider Electric EcoStruxure/Building Operation 2.0/Enterprise Server/db

const (
	key_system_services = `SYSTEM\CurrentControlSet\Services`
	es_backup_folder    = `db_backup`
)

type eboService struct {
	key   string
	name  string
	image string
}

func listLocations() {

	ess, err := EnterpriseServers()
	if err != nil {
		return
	}

	for _, es := range ess {
		dbPath, _ := es.DBPath()
		fmt.Println(es.name, ":", dbPath)
	}

}

func EnterpriseServers() ([]*eboService, error) {

	k, err := registry.OpenKey(registry.LOCAL_MACHINE, key_system_services, registry.ENUMERATE_SUB_KEYS)
	if err != nil {
		return nil, err
	}
	defer k.Close()

	services, err := k.ReadSubKeyNames(0)
	if err != nil {
		return nil, err
	}

	eboServices := make([]*eboService, 0)
	n := 0
	for i := range services {
		if strings.HasSuffix(services[i], "Enterprise Server") {
			key := fmt.Sprintf(`%s\%s`, key_system_services, services[i])
			s, err := readService(key)
			if err != nil {
				continue
			}
			eboServices = append(eboServices, s)
			n++
		}
	}
	return eboServices, nil
}

func readService(key string) (*eboService, error) {

	k, err := registry.OpenKey(registry.LOCAL_MACHINE, key, registry.QUERY_VALUE)
	if err != nil {
		return nil, err
	}
	defer k.Close()

	displayName, _, err := k.GetStringValue("DisplayName")
	if err != nil {
		return nil, err
	}

	imagePath, _, err := k.GetStringValue("ImagePath")
	if err != nil {
		return nil, err
	}

	eboService1 := &eboService{
		key:   key,
		name:  displayName,
		image: trimQuote(imagePath),
	}

	return eboService1, nil
}

func trimQuote(s string) string {
	if s[0] == '"' {
		s = s[1:]
	}
	if s[len(s)-1] == '"' {
		s = s[:len(s)-1]
	}
	return s
}

func readImagePath(key string) (string, error) {

	k, err := registry.OpenKey(registry.LOCAL_MACHINE, key, registry.QUERY_VALUE)
	if err != nil {
		return "", err
	}
	defer k.Close()

	s, _, err := k.GetStringValue("ImagePath")
	if err != nil {
		return "", err
	}

	return trimQuote(s), nil
}

var ErrNotFound = errors.New("path to DB folder not found")

func readDBPath(installPath string) (string, error) {

	// C:\Program Files (x86)\Schneider Electric EcoStruxure\Building Operation 2.0\Enterprise Server\etc\dbpath.properties
	// server.paths.db=C:/ProgramData/Schneider Electric EcoStruxure/Building Operation 2.0/Enterprise Server/db

	pf := filepath.Join(installPath, `etc\dbpath.properties`)
	f, err := os.ReadFile(pf)
	if err != nil {
		// fmt.Println(err)
		return "", err
	}

	s := bufio.NewScanner(bytes.NewReader(f))
	for s.Scan() {
		kv := strings.Split(s.Text(), "=")
		if len(kv) != 2 {
			continue
		}
		k := strings.ToLower(strings.TrimSpace(kv[0]))
		if k != "server.paths.db" {

			continue
		}
		v := strings.TrimSpace(kv[1])
		return filepath.Dir(trimQuote(v)), nil
	}
	return "", ErrNotFound
}

func EnterpriseServersPaths(es_service_keys []string) []string {

	installFolders := make([]string, len(es_service_keys))
	for i, key := range es_service_keys {
		imgPath, err := readImagePath(key)
		if err != nil {
			continue
		}
		installFolders[i] = filepath.Dir(filepath.Dir(imgPath))
	}

	return installFolders
}

func EnterpriseServersDBPaths(paths []string) []string {

	dbFolders := make([]string, 0, len(paths))
	for i := range paths {
		img, err := readDBPath(paths[i])
		if err != nil {
			continue
		}
		dbFolders = append(dbFolders, img)
	}

	return dbFolders
}

func (es *eboService) InstallPath() string {
	return filepath.Dir(filepath.Dir(es.image))

}

func (es *eboService) DBPath() (string, error) {
	pf := filepath.Join(es.InstallPath(), `etc\dbpath.properties`)

	f, err := os.ReadFile(pf)
	if err != nil {
		return "", err
	}

	s := bufio.NewScanner(bytes.NewReader(f))
	for s.Scan() {
		kv := strings.Split(s.Text(), "=")
		if len(kv) != 2 {
			continue
		}
		k := strings.ToLower(strings.TrimSpace(kv[0]))
		if k != "server.paths.db" {
			continue
		}
		return strings.TrimSpace(kv[1]), nil
	}
	return "", ErrNotFound
}
