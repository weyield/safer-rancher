package main

import (
	"github.com/rancher/go-rancher/v2"
	"time"
	"fmt"
	"strings"
)

type Rancher struct {
	client *client.RancherClient
}

type RancherServiceInfo struct {
	rancherId string
	rancherName string
	dockerImage string
	hostname string
}

func (postgresInfo RancherServiceInfo) String() string {
	return fmt.Sprintf("%s (id: %s, name: %s, image: %s)",
		postgresInfo.hostname, postgresInfo.rancherId, postgresInfo.rancherName, postgresInfo.dockerImage)
}

type ServiceIpNotFoundError struct {
	ServiceId string
	ServiceName string
}

func (e *ServiceIpNotFoundError) Error() string {
	return fmt.Sprintf("Cannot find an IP address for the service named %s (id: %s)", e.ServiceName, e.ServiceId)
}

type RancherServerServiceError struct {}

func (e *RancherServerServiceError) Error() string {
	return fmt.Sprintf("Cannot find the service of the Rancher server. The host of the server must be added to Rancher.")
}

func NewRancher(url, accessKey, secretKey string) (*Rancher, error) {
	var options = client.ClientOpts{url,accessKey,secretKey,time.Minute * 5}
	rancherClient, err := client.NewRancherClient(&options)
	if err != nil {
		return nil, err
	}
	rancher := Rancher{rancherClient}
	return &rancher, nil
}

func findServiceHostname(rancherClient *client.RancherClient, service client.Service) (string, error) {
	var instanceCollection client.ContainerCollection
	if err := rancherClient.GetLink(service.Resource, "instances", &instanceCollection); err != nil {
		return "", err
	}
	for _, instance := range instanceCollection.Data {
		if instance.State == "running" && instance.PrimaryIpAddress != "" {
			return instance.PrimaryIpAddress, nil
		}
	}
	return "", &ServiceIpNotFoundError{service.Name, service.Id}
}

func (r *Rancher) getPostgresServices() ([]*RancherServiceInfo, error) {
	projectClient := r.client.Project
	projects, err := projectClient.List(emptyListOpts)
	if err != nil {
		return nil, err
	}
	pgServices := make([]*RancherServiceInfo, 0)
	for _, project := range projects.Data {
		var serviceCollection client.ServiceCollection
		if err = r.client.GetLink(project.Resource, "services", &serviceCollection); err != nil {
			return nil, err
		}

		for _, service := range serviceCollection.Data {
			imageUuid := service.LaunchConfig.ImageUuid
			if strings.Count(imageUuid, "postgres") > 0 && service.State == "active" {
				hostname, err := findServiceHostname(r.client, service)
				if err != nil {
					return nil, err
				}
				pgService := RancherServiceInfo{service.Id, service.Name, strings.Replace(service.LaunchConfig.ImageUuid, "docker:", "", 1), hostname}
				pgServices = append(pgServices, &pgService)
			}
		}
	}
	return pgServices, nil
}

func (r *Rancher) getRancherServerService() (*RancherServiceInfo, error) {
	projectClient := r.client.Project
	projects, err := projectClient.List(emptyListOpts)
	if err != nil {
		return nil, err
	}
	for _, project := range projects.Data {
		var containerCollection client.ContainerCollection
		if err = r.client.GetLink(project.Resource, "containers", &containerCollection); err != nil {
			return nil, err
		}

		for _, container := range containerCollection.Data {
			imageUuid := container.ImageUuid
			if strings.Count(imageUuid, "rancher/server") > 0 {
				rancherServerService := RancherServiceInfo{container.Id, container.Name, strings.Replace(container.ImageUuid, "docker:", "", 1), container.Hostname}
				return &rancherServerService, nil
			}
		}
	}
	return nil, &RancherServerServiceError{}
}

var emptyListOpts *client.ListOpts = &client.ListOpts{make(map[string]interface{})}
