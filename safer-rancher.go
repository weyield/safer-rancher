package main

import (
	"fmt"
	"os"
)

var (
	backupRoot string
	rancherUrl string
	rancherAccessKey string
	rancherSecretKey string
)

func main() {
	var ok bool
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

	rancher, err := NewRancher(rancherUrl, rancherAccessKey, rancherSecretKey)
	if err != nil {
		fmt.Println("An error occured during the rancher API connection:\n" + err.Error())
		return
	}
	postgresInfos, err := rancher.getPostgresServices()
	if err != nil {
		fmt.Println("An error occured during postgres services retrieval:\n" + err.Error())
		return
	}

	rancherServerInfo, err := rancher.getRancherServerService()
	if err != nil {
		fmt.Println("An error occured during rancher server service retrieval:\n" + err.Error())
		return
	}
	fmt.Println("Starting backup of rancher database")
	if err := dumpRancherDatabase(rancherServerInfo.hostname, backupRoot); err != nil {
		fmt.Println("An error occured while dumping the rancher database:\n" + err.Error())
		return
	}

	for i, postgresInfo := range postgresInfos {
		fmt.Printf("Starting backup of postgres service %d/%d: %s\n", i+1, len(postgresInfos), postgresInfo.rancherName)
		if err := dumpPostgresDatabase(postgresInfo, backupRoot); err != nil {
			fmt.Println("An error occured while dumping a postgres service:\n" + err.Error())
			return
		}
	}

	fmt.Println("Backup finished")
}
