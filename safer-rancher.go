package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

var (
	backupRoot               string
	rancherUrl               string
	rancherAccessKey         string
	rancherSecretKey         string
	activateRancherDatabase  bool
	activatePostgresDatabase bool
)

func main() {
	var ok bool
	var activate string
	if backupRoot, ok = os.LookupEnv("BACKUP_ROOT"); !ok {
		panic("The BACKUP_ROOT environment variable is required")
	}
	if rancherUrl, ok = os.LookupEnv("RANCHER_URL"); !ok {
		panic("The RANCHER_URL environment variable is required")
	}
	if rancherAccessKey, ok = os.LookupEnv("RANCHER_ACCESS_KEY"); !ok {
		panic("The RANCHER_ACCESS_KEY environment variable is required")
	}
	if rancherSecretKey, ok = os.LookupEnv("RANCHER_SECRET_KEY"); !ok {
		panic("The RANCHER_SECRET_KEY environment variable is required")
	}

	if activate, ok = os.LookupEnv("ACTIVATE_RANCHER_DATABASE"); ok {
		activateRancherDatabase = strings.ToLower(activate) != "false"
	} else {
		// Default value
		activateRancherDatabase = true
	}
	if activate, ok = os.LookupEnv("ACTIVATE_POSTGRES_DATABASE"); ok {
		activatePostgresDatabase = strings.ToLower(activate) != "false"
	} else {
		// Default value
		activatePostgresDatabase = true
	}

	rancher, err := NewRancher(rancherUrl, rancherAccessKey, rancherSecretKey)
	if err != nil {
		fmt.Println("An error occured during the rancher API connection:\n" + err.Error())
		return
	}

	if activateRancherDatabase {
		rancherServerInfo, err := rancher.getRancherServerService()
		if err != nil {
			fmt.Println("An error occured during rancher server service retrieval:\n" + err.Error())
			return
		}
		fmt.Println("Starting backup of rancher database")
		if err := dumpRancherDatabase(rancherServerInfo.ip, backupRoot); err != nil {
			fmt.Println("An error occured while dumping the rancher database:\n" + err.Error())
			return
		}
	}

	if activatePostgresDatabase {
		backupLog := ""
		postgresInfos, err := rancher.getPostgresServices()
		if err != nil {
			fmt.Println("An error occured during postgres services retrieval:\n" + err.Error())
			return
		}
		for i, postgresInfo := range postgresInfos {
			fmt.Printf("Starting backup of postgres service %d/%d: %s\n", i+1, len(postgresInfos), postgresInfo.rancherName)
			dumpPath := fmt.Sprintf("/backup/dump_%s_%s.sql", postgresInfo.rancherId, postgresInfo.rancherName)
			if err := dumpPostgresDatabase(postgresInfo, backupRoot, dumpPath); err != nil {
				fmt.Println("An error occured while dumping a postgres service:\n" + err.Error())
				return
			}
			backupLog += postgresInfo.String() + "\n"
		}
		if err := ioutil.WriteFile(backupRoot+"/backuplog.csv", []byte(backupLog), 0644); err != nil {
			fmt.Println("An error occured while writing the log file:\n" + err.Error())
			return
		}
	}

	fmt.Println("Backup finished")
}
