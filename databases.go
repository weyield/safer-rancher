package main

import (
	"fmt"
	"github.com/docker/docker/client"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"golang.org/x/net/context"
	"io"
	"io/ioutil"
	"os"
	"strings"
)

type dumpError struct {
	dockerImage string
	dumpCommand string
}

func (e dumpError) Error() string {
	return fmt.Sprintf("Dump command failed: %s (using %s)", e.dumpCommand, e.dockerImage)
}

func makeDump(dumpCommand string, backupPath string, dockerImage string, links []string) error {

	ctx := context.Background()

	cli, err := client.NewEnvClient()
	if err != nil {
		return err
	}

	pullOut, err := cli.ImagePull(ctx, dockerImage, types.ImagePullOptions{})
	if err != nil {
		return err
	}
	if _, err = ioutil.ReadAll(pullOut); err != nil {
		return err
	}

	newDumpCommand := []string{"bash", "-c", "sleep 5 && " + dumpCommand}
	// Sleep 5 seconds in order to let rancher configure the network

	containerConfig := &container.Config{
		Image: dockerImage,
		Cmd:   newDumpCommand,
		Labels: map[string]string{"io.rancher.container.network": "true"},
	}

	hostConfig := &container.HostConfig{
		AutoRemove: true,
		Binds: []string{backupPath+":/backup"},
		Links: links,
	}

	resp, err := cli.ContainerCreate(ctx, containerConfig, hostConfig, nil, "")
	if err != nil {
		return err
	}

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return err
	}

	statusCodeChan, errChan := cli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	if len(errChan) > 0 {
		err = <-errChan
		return err
	}

	out, err := cli.ContainerLogs(ctx, resp.ID, types.ContainerLogsOptions{ShowStdout: true, ShowStderr: true})
	if err != nil {
		return err
	}
	io.Copy(os.Stdout, out)

	statusCode := <- statusCodeChan
	if statusCode.StatusCode != 0 {
		return dumpError{dockerImage, strings.Join(newDumpCommand, " ")}
	}

	return nil
}

func dumpPostgresDatabase(info *RancherServiceInfo, backupPath string) error {
	dumpPath := fmt.Sprintf("/backup/dump_%s_%s.sql", info.rancherId, info.rancherName)
	dumpCommand := fmt.Sprintf("pg_dumpall -h %s -U postgres --clean -f %s", info.hostname, dumpPath)

	if err := makeDump(dumpCommand, backupPath, info.dockerImage, []string{}); err != nil {
		return err
	}

	return nil
}

func dumpRancherDatabase(rancherContainerId, backupPath string) error {
	dumpPath := "/backup/rancher_dump.sql"
	dumpCommand := "mysqldump -A -h db -u cattle -pcattle --result-file " + dumpPath

	if err := makeDump(dumpCommand, backupPath, "mysql", []string{rancherContainerId+":db"}); err != nil {
		return err
	}

	return nil
}
