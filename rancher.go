package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/rancher/go-rancher/v2"
)

type Rancher struct {
	client *client.RancherClient
}

type RancherServiceInfo struct {
	rancherId   string
	rancherName string
	dockerImage string
	ip          string
	hostname    string
}

func (postgresInfo RancherServiceInfo) String() string {
	return fmt.Sprintf("%s,%s,%s,%s,%s",
		postgresInfo.rancherId, postgresInfo.rancherName, postgresInfo.dockerImage, postgresInfo.ip, postgresInfo.hostname)
}

type ServiceIpNotFoundError struct {
	ServiceId   string
	ServiceName string
}

func (e *ServiceIpNotFoundError) Error() string {
	return fmt.Sprintf("Cannot find an IP address for the service named %s (id: %s)", e.ServiceName, e.ServiceId)
}

type RancherServerServiceError struct{}

func (e *RancherServerServiceError) Error() string {
	return fmt.Sprintf("Cannot find the service of the Rancher server. The host of the server must be added to Rancher.")
}

func NewRancher(url, accessKey, secretKey string) (*Rancher, error) {
	var options = client.ClientOpts{Url: url, AccessKey: accessKey, SecretKey: secretKey, Timeout: time.Minute * 5}
	rancherClient, err := client.NewRancherClient(&options)
	if err != nil {
		return nil, err
	}
	rancher := Rancher{rancherClient}
	return &rancher, nil
}

func findServiceHostname(rancherClient *client.RancherClient, service client.Service) (string, string, error) {
	var instanceCollection client.ContainerCollection
	if err := rancherClient.GetLink(service.Resource, "instances", &instanceCollection); err != nil {
		return "", "", err
	}
	for _, instance := range instanceCollection.Data {
		if instance.State == "running" && instance.PrimaryIpAddress != "" {
			var hostCollection client.HostCollection
			if err := rancherClient.GetLink(instance.Resource, "hosts", &hostCollection); err != nil {
				return "", "", err
			}
			for _, host := range hostCollection.Data {
				return instance.PrimaryIpAddress, host.Hostname, nil
			}
		}
	}
	return "", "", &ServiceIpNotFoundError{service.Name, service.Id}
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
				ip, hostname, err := findServiceHostname(r.client, service)
				if err != nil {
					return nil, err
				}
				pgService := RancherServiceInfo{service.Id, service.Name, strings.Replace(service.LaunchConfig.ImageUuid, "docker:", "", 1), ip, hostname}
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
				rancherServerService := RancherServiceInfo{container.Id, container.Name, strings.Replace(container.ImageUuid, "docker:", "", 1), container.Hostname, ""}
				return &rancherServerService, nil
			}
		}
	}
	return nil, &RancherServerServiceError{}
}

var emptyListOpts *client.ListOpts = &client.ListOpts{Filters: make(map[string]interface{})}
