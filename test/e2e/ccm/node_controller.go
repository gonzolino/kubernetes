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

package ccm

import (
	"fmt"
	"time"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubeletapis "k8s.io/kubernetes/pkg/kubelet/apis"
	"k8s.io/kubernetes/test/e2e/ccm/cloudclient"
	"k8s.io/kubernetes/test/e2e/framework"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func waitForNode(f *framework.Framework, nodeInfo *cloudclient.NodeInfo) (*v1.Node, error) {
	nodeClient := f.ClientSet.CoreV1().Nodes()

	nodes, err := nodeClient.List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	for _, n := range nodes.Items {
		if n.Name == nodeInfo.Name {
			return &n, nil
		}
	}
	return nil, fmt.Errorf("Could not find node %s", nodeInfo.Name)
}

var _ = SIGDescribe("NodeController [Feature:CCM]", func() {
	f := framework.NewDefaultFramework("node-controller")
	cloud, err := cloudclient.NewCloud("openstack")
	if err != nil {
		framework.Failf("Failed to initialize cloud client: %s", err)
	}
	var nodeInfo *cloudclient.NodeInfo
	var node *v1.Node

	BeforeEach(func() {
		// TODO: randomize name
		nodeInfo, err = cloud.CreateNode("ccm-test")
		if err != nil {
			framework.Failf("Could not create test node: %s", err)
		}
		framework.Logf("Created node %s for node controller tests", nodeInfo.ProviderID)
		time.Sleep(60 * time.Second)
		node, err = waitForNode(f, nodeInfo)
		if err != nil {
			framework.Failf("Failed to find created node: %s", err)
		}
		framework.Logf("Node %s (provider ID: %s)is now registered with teh kubernetes cluster", node.Name, nodeInfo.ProviderID)
	})

	AfterEach(func() {
		err = cloud.DeleteNode(nodeInfo)
		if err != nil {
			framework.Failf("Failed to delete test node: %s", err)
		}
		framework.Logf("Deleted node %s", nodeInfo.ProviderID)
	})

	It("should initialize a node", func() {
		labels := make([]string, 0, len(node.Labels))
		for label := range node.Labels {
			labels = append(labels, label)
		}

		By("with cloud specific provider ID")
		Expect(node.Spec.ProviderID).To(Equal(nodeInfo.ProviderID))
		framework.Logf("Node %s provider ID: %s", node.Name, node.Spec.ProviderID)

		By("with cloud specific instance type label")
		Expect(labels).Should(ContainElement(kubeletapis.LabelInstanceType))
		Expect(node.Labels[kubeletapis.LabelInstanceType]).To(Equal(nodeInfo.InstanceType))
		framework.Logf("Node %s instance type: %s", node.Name, node.Labels[kubeletapis.LabelInstanceType])

		//By("with cloud specific region labels")
		//Expect(labels).Should(ContainElement(kubeletapis.LabelZoneRegion))
		//Expect(labels).Should(ContainElement(kubeletapis.LabelZoneFailureDomain))

		By("with the cloud provided address / hostname")
		Expect(node.Status.Addresses).NotTo(BeEmpty())
		framework.Logf("Node %s address: %s", node.Name, node.Status.Addresses[0])
	})

	// It("should update the node addresses upon restarts", func() {
	// 	// TODO: dummy
	// 	Expect(1).To(Equal(1))
	// })
})
