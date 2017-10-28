/*
Copyright 2016 The Kubernetes Authors.

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

package gce

import (
	"fmt"
	"strings"

	"github.com/golang/glog"
	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	compute "google.golang.org/api/compute/v0.beta"
	"google.golang.org/api/iam/v1"
	"google.golang.org/api/storage/v1"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kubernetes/federation/pkg/dnsprovider"
	"k8s.io/kubernetes/federation/pkg/dnsprovider/providers/google/clouddns"
)

type GCECloud interface {
	fi.Cloud
	IAM() *iam.Service
	Compute() *compute.Service
	Storage() *storage.Service

	Region() string
	Project() string
	WaitForOp(op *compute.Operation) error
	GetApiIngressStatus(cluster *kops.Cluster) ([]kops.ApiIngressStatus, error)
	Labels() map[string]string

	// FindClusterStatus gets the status of the cluster as it exists in GCE, inferred from volumes
	FindClusterStatus(cluster *kops.Cluster) (*kops.ClusterStatus, error)

	Zones() ([]string, error)
}

type gceCloudImplementation struct {
	compute *compute.Service
	iam     *iam.Service
	storage *storage.Service

	region  string
	project string

	labels map[string]string
}

var _ fi.Cloud = &gceCloudImplementation{}

func (c *gceCloudImplementation) ProviderID() kops.CloudProviderID {
	return kops.CloudProviderGCE
}

var gceCloudInstances map[string]GCECloud = make(map[string]GCECloud)

func NewGCECloud(region string, project string, labels map[string]string) (GCECloud, error) {
	i := gceCloudInstances[region+"::"+project]
	if i != nil {
		return i.(gceCloudInternal).WithLabels(labels), nil
	}

	c := &gceCloudImplementation{region: region, project: project}

	ctx := context.Background()

	client, err := google.DefaultClient(ctx, compute.ComputeScope)
	if err != nil {
		return nil, fmt.Errorf("error building google API client: %v", err)
	}
	computeService, err := compute.New(client)
	if err != nil {
		return nil, fmt.Errorf("error building compute API client: %v", err)
	}
	c.compute = computeService

	iamService, err := iam.New(client)
	if err != nil {
		return nil, fmt.Errorf("error building iam API client: %v", err)
	}
	c.iam = iamService

	storageService, err := storage.New(client)
	if err != nil {
		return nil, fmt.Errorf("error building storage API client: %v", err)
	}
	c.storage = storageService

	gceCloudInstances[region+"::"+project] = c

	return c.WithLabels(labels), nil
}

// gceCloudInternal is an interface for private functions for a gceCloudImplemention or mockGCECloud
type gceCloudInternal interface {
	// WithLabels returns a copy of the GCECloud, bound to the specified labels
	WithLabels(labels map[string]string) GCECloud
}

// WithLabels returns a copy of the GCECloud, bound to the specified labels
func (c *gceCloudImplementation) WithLabels(labels map[string]string) GCECloud {
	i := &gceCloudImplementation{}
	*i = *c
	i.labels = labels
	return i
}

// Compute returns private struct element compute.
func (c *gceCloudImplementation) Compute() *compute.Service {
	return c.compute
}

// IAM returns the IAM service client.
func (c *gceCloudImplementation) IAM() *iam.Service {
	return c.iam
}

// Storage returns private struct element storage.
func (c *gceCloudImplementation) Storage() *storage.Service {
	return c.storage
}

// Region returns private struct element region.
func (c *gceCloudImplementation) Region() string {
	return c.region
}

// Project returns private struct element project.
func (c *gceCloudImplementation) Project() string {
	return c.project
}

func (c *gceCloudImplementation) DNS() (dnsprovider.Interface, error) {
	provider, err := clouddns.CreateInterface(c.project, nil)
	if err != nil {
		return nil, fmt.Errorf("Error building (k8s) DNS provider: %v", err)
	}
	return provider, nil
}

func (c *gceCloudImplementation) FindVPCInfo(id string) (*fi.VPCInfo, error) {
	glog.Warningf("FindVPCInfo not (yet) implemented on GCE")
	return nil, nil
}

func (c *gceCloudImplementation) Labels() map[string]string {
	// Defensive copy
	tags := make(map[string]string)
	for k, v := range c.labels {
		tags[k] = v
	}
	return tags
}

// TODO refactor this out of resources
// this is needed for delete groups and other new methods

// Zones returns the zones in a region
func (c *gceCloudImplementation) Zones() ([]string, error) {

	var zones []string
	// TODO: Only zones in api.Cluster object, if we have one?
	gceZones, err := c.Compute().Zones.List(c.Project()).Do()
	if err != nil {
		return nil, fmt.Errorf("error listing zones: %v", err)
	}
	for _, gceZone := range gceZones.Items {
		u, err := ParseGoogleCloudURL(gceZone.Region)
		if err != nil {
			return nil, err
		}
		if u.Name != c.Region() {
			continue
		}
		zones = append(zones, gceZone.Name)
	}
	if len(zones) == 0 {
		return nil, fmt.Errorf("unable to determine zones in region %q", c.Region())
	}

	glog.Infof("Scanning zones: %v", zones)
	return zones, nil
}

func (c *gceCloudImplementation) WaitForOp(op *compute.Operation) error {
	return WaitForOp(c.compute, op)
}

func (c *gceCloudImplementation) GetApiIngressStatus(cluster *kops.Cluster) ([]kops.ApiIngressStatus, error) {
	var ingresses []kops.ApiIngressStatus

	// Note that this must match GCEModelContext::NameForForwardingRule
	name := SafeObjectName("api", cluster.ObjectMeta.Name)

	glog.V(2).Infof("Querying GCE to find ForwardingRules for API (%q)", name)
	forwardingRule, err := c.compute.ForwardingRules.Get(c.project, c.region, name).Do()
	if err != nil {
		if !IsNotFound(err) {
			forwardingRule = nil
		} else {
			return nil, fmt.Errorf("error getting ForwardingRule %q: %v", name, err)
		}
	}

	if forwardingRule != nil {
		if forwardingRule.IPAddress == "" {
			return nil, fmt.Errorf("Found forward rule %q, but it did not have an IPAddress", name)
		}

		ingresses = append(ingresses, kops.ApiIngressStatus{
			IP: forwardingRule.IPAddress,
		})
	}

	return ingresses, nil
}

// FindInstanceTemplates finds all instance templates that are associated with the current cluster
// It matches them by looking for instance metadata with key='cluster-name' and value of our cluster name
func FindInstanceTemplates(c GCECloud, clusterName string) ([]*compute.InstanceTemplate, error) {
	findClusterName := strings.TrimSpace(clusterName)
	var matches []*compute.InstanceTemplate
	ctx := context.Background()

	err := c.Compute().InstanceTemplates.List(c.Project()).Pages(ctx, func(page *compute.InstanceTemplateList) error {
		for _, t := range page.Items {
			match := false
			for _, item := range t.Properties.Metadata.Items {
				if item.Key == "cluster-name" {
					value := fi.StringValue(item.Value)
					if strings.TrimSpace(value) == findClusterName {
						match = true
					} else {
						match = false
						break
					}
				}
			}

			if !match {
				continue
			}

			matches = append(matches, t)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("error listing instance templates: %v", err)
	}

	return matches, nil
}
