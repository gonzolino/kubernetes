/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cloudclient

import (
	"fmt"
	"os"
	"strings"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/servers"
	"github.com/gophercloud/gophercloud/pagination"
)

type openstackCloud struct {
	provider *gophercloud.ProviderClient
	region   string
	clients  map[string]*gophercloud.ServiceClient
}

func newOpenstackCloud() (*openstackCloud, error) {
	// TODO: Authenticate from test context not from env vars
	opts, err := openstack.AuthOptionsFromEnv()
	if err != nil {
		return nil, err
	}

	provider, err := openstack.AuthenticatedClient(opts)
	if err != nil {
		return nil, fmt.Errorf("Failed to create OpenStack client: %s", err)
	}

	return &openstackCloud{
		provider: provider,
		region:   os.Getenv("OS_REGION"),
		clients:  make(map[string]*gophercloud.ServiceClient),
	}, nil
}

func (cloud *openstackCloud) getServiceClient(service string) (*gophercloud.ServiceClient, error) {
	if client, ok := cloud.clients[service]; ok {
		return client, nil
	}
	opts := gophercloud.EndpointOpts{Region: cloud.region}
	var client *gophercloud.ServiceClient
	var err error
	switch service {
	case "compute":
		client, err = openstack.NewComputeV2(cloud.provider, opts)
	}
	if err != nil {
		return nil, err
	}
	cloud.clients[service] = client
	return client, nil
}

func (cloud *openstackCloud) getNode(name string) (*NodeInfo, error) {
	compute, err := cloud.getServiceClient("compute")
	if err != nil {
		return nil, fmt.Errorf("Failed to create compute client: %s", err)
	}

	pager := servers.List(compute, servers.ListOpts{})
	var server servers.Server
	err = pager.EachPage(func(page pagination.Page) (bool, error) {
		serverList, err := servers.ExtractServers(page)
		if err != nil {
			return false, err
		}
		for _, s := range serverList {
			if s.Name == name {
				server = s
				return false, nil
			}
		}
		return true, nil
	})
	if err != nil {
		return nil, err
	}

	return &NodeInfo{
		Name:         server.Name,
		ProviderID:   "openstack:///" + server.ID,
		InstanceType: server.Flavor["id"].(string),
	}, nil
}

func (cloud *openstackCloud) createNode(name string) (*NodeInfo, error) {
	compute, err := cloud.getServiceClient("compute")
	if err != nil {
		return nil, fmt.Errorf("Failed to create compute client: %s", err)
	}

	// TODO: Configure CreateOpts from test context, not from env vars
	flavorID := os.Getenv("OS_TEST_FLAVOR")
	imageID := os.Getenv("OS_TEST_IMAGE")
	networkID := os.Getenv("OS_TEST_NETWORK")
	securityGroups := strings.Split(os.Getenv("OS_TEST_SECGROUPS"), ",")
	userData := []byte(os.Getenv("OS_TEST_USER_DATA"))

	network := servers.Network{
		UUID: networkID,
	}

	createOpts := servers.CreateOpts{
		Name:           name,
		FlavorRef:      flavorID,
		ImageRef:       imageID,
		SecurityGroups: securityGroups,
		UserData:       userData,
		Networks:       []servers.Network{network},
	}
	server, err := servers.Create(compute, createOpts).Extract()
	if err != nil {
		return nil, err
	}
	server, err = servers.Get(compute, server.ID).Extract()
	if err != nil {
		return nil, err
	}

	return &NodeInfo{
		Name:         server.Name,
		ProviderID:   "openstack:///" + server.ID,
		InstanceType: server.Flavor["id"].(string),
	}, nil
}

func (cloud *openstackCloud) deleteNode(node *NodeInfo) error {
	compute, err := cloud.getServiceClient("compute")
	if err != nil {
		return fmt.Errorf("Failed to create compute client: %s", err)
	}

	// Get server id by cutting of 'openstack:///' from providerID
	id := node.ProviderID[13:]
	return servers.Delete(compute, id).ExtractErr()
}
