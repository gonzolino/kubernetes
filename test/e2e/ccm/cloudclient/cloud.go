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

	"k8s.io/kubernetes/test/e2e/framework"
)

// NodeInfo contains information about a cloud node
type NodeInfo struct {
	Name         string
	ProviderID   string
	InstanceType string
}

// Cloud is a representation of the tested cloud
type Cloud struct {
	cloudType string
	// TODO: add other clouds
	openstack *openstackCloud
}

// NewCloud creates a Cloud object that acts as an abstraction to the cloud underlying the k8s cluster.
func NewCloud(cloudType string) (*Cloud, error) {
	// TODO: Instead of passing in a string to determine the cloud type,
	// we must identify the correct cloud client to use based on the test context provider.
	switch cloudType {
	case "openstack":
		cloud, err := newOpenstackCloud()
		if err != nil {
			return nil, err
		}
		return &Cloud{
			cloudType: "openstack",
			openstack: cloud,
		}, nil
	default:
		return nil, fmt.Errorf("%s is an unsupported cloud provider", framework.TestContext.Provider)
	}
}

// GetNode retrieves information about the node with the given name from the cloud.
func (cloud *Cloud) GetNode(name string) (*NodeInfo, error) {
	switch cloud.cloudType {
	case "openstack":
		return cloud.openstack.getNode(name)
	default:
		return nil, fmt.Errorf("%s is an unsupported provider", cloud.cloudType)
	}
}

// CreateNode creates a node in the cloud and returns information for that node
func (cloud *Cloud) CreateNode(name string) (*NodeInfo, error) {
	switch cloud.cloudType {
	case "openstack":
		return cloud.openstack.createNode(name)
	default:
		return nil, fmt.Errorf("%s is an unsuported provider", cloud.cloudType)
	}
}

// DeleteNode deletes a node in the cloud
func (cloud *Cloud) DeleteNode(node *NodeInfo) error {
	switch cloud.cloudType {
	case "openstack":
		return cloud.openstack.deleteNode(node)
	default:
		return fmt.Errorf("%s is an unsuported provider", cloud.cloudType)
	}
}
