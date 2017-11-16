# Safer Rancher
Backup script for rancher written in GO. Currently, Safer Rancher will backup the Rancher's MySQL database and all the PostgreSQL instances managed by Rancher. The host hosting the Rancher server must be added to the Rancher's hosts. There are 3 ways of using this script: from our rancher catalog, with docker, or using the binary program.

## Usage with our catalog (prefered)
The easiest way of using Safer Rancher is to install it from our catalog. You can find the install instructions on the [catalog's repository](https://github.com/weyield/weyield-rancher-catalog).

## Usage with Docker
The Docker image is built automatically by Docker Hub. You can use it like this:
```
run -e BACKUP_ROOT="/root/backup" -e RANCHER_URL="https://xxx/v2-beta" -e RANCHER_ACCESS_KEY="XXX" -e RANCHER_SECRET_KEY="XxXxXx" -v /var/run/docker.sock:/var/run/docker.sock -v /root/backup:/root/backup safer-rancher
```

## Usage with the Go binary
### Compilation
* Clone the repository in `$GOROOT/github.com/weyield/`
* go to the `rancher-safer` directory
* run `go get .`
* run `go build -o safer-rancher .`
### Start
The `safer-rancher` command need to access the rancher network. Therefore, it must be started on a host controlled by the Rancher instance you want to backup.
You must define the following environment variables:
* `BACKUP_ROOT`: Full path of the directory where the backups are saved
* `RANCHER_URL`: Url of the Rancher API, such as `https://xxx/v2-beta`
* `RANCHER_ACCESS_KEY`: Access key of the Rancher API
* `RANCHER_SECRET_KEY`: Secret key of the Rancher API
For example, you can run:
```
BACKUP_ROOT="/root/backup" RANCHER_URL="https://xxx/v2-beta" RANCHER_ACCESS_KEY="XXX" RANCHER_SECRET_KEY="XxXxXx" ./safer-rancher
```

Copyright (c) 2017 WeYield
