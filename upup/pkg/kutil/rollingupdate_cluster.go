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

package kutil

import (
	"fmt"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/golang/glog"
	api "k8s.io/kops/pkg/apis/kops"
	validate "k8s.io/kops/pkg/validation"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kubernetes/pkg/api/v1"
	"k8s.io/kubernetes/pkg/client/clientset_generated/release_1_5"
)

// RollingUpdateCluster restarts cluster nodes
type RollingUpdateCluster struct {
	Cloud fi.Cloud

	MasterInterval time.Duration
	NodeInterval   time.Duration

	Force bool
}

// TODO move retries to RollingUpdateCluster
const retries = 8

// Find CloudInstanceGroups
func FindCloudInstanceGroups(cloud fi.Cloud, cluster *api.Cluster, instancegroups []*api.InstanceGroup, warnUnmatched bool, nodes []v1.Node) (map[string]*CloudInstanceGroup, error) {
	awsCloud := cloud.(awsup.AWSCloud)

	groups := make(map[string]*CloudInstanceGroup)

	tags := awsCloud.Tags()

	asgs, err := findAutoscalingGroups(awsCloud, tags)
	if err != nil {
		return nil, err
	}

	nodeMap := make(map[string]*v1.Node)
	for i := range nodes {
		node := &nodes[i]
		awsID := node.Spec.ExternalID
		nodeMap[awsID] = node
	}

	for _, asg := range asgs {
		name := aws.StringValue(asg.AutoScalingGroupName)
		var instancegroup *api.InstanceGroup
		for _, g := range instancegroups {
			var asgName string
			switch g.Spec.Role {
			case api.InstanceGroupRoleMaster:
				asgName = g.Name + ".masters." + cluster.Name
			case api.InstanceGroupRoleNode:
				asgName = g.Name + "." + cluster.Name
			default:
				glog.Warningf("Ignoring InstanceGroup of unknown role %q", g.Spec.Role)
				continue
			}

			if name == asgName {
				if instancegroup != nil {
					return nil, fmt.Errorf("Found multiple instance groups matching ASG %q", asgName)
				}
				instancegroup = g
			}
		}
		if instancegroup == nil {
			if warnUnmatched {
				glog.Warningf("Found ASG with no corresponding instance group: %q", name)
			}
			continue
		}
		group := buildCloudInstanceGroup(instancegroup, asg, nodeMap)
		groups[instancegroup.Name] = group
	}

	return groups, nil
}

// Perform a rolling update on a K8s Cluster
func (c *RollingUpdateCluster) RollingUpdate(groups map[string]*CloudInstanceGroup, instanceGroups *api.InstanceGroupList, k8sClient *release_1_5.Clientset) error {
	if len(groups) == 0 {
		return nil
	}

	var resultsMutex sync.Mutex
	results := make(map[string]error)

	masterGroups := make(map[string]*CloudInstanceGroup)
	nodeGroups := make(map[string]*CloudInstanceGroup)
	for k, group := range groups {
		switch group.InstanceGroup.Spec.Role {
		case api.InstanceGroupRoleNode:
			nodeGroups[k] = group
		case api.InstanceGroupRoleMaster:
			masterGroups[k] = group
		default:
			return fmt.Errorf("unknown group type for group %q", group.InstanceGroup.Name)
		}
	}

	// Upgrade master first
	{
		var wg sync.WaitGroup

		// We run master nodes in series, even if they are in separate instance groups
		// typically they will be in separate instance groups, so we can force the zones,
		// and we don't want to roll all the masters at the same time.  See issue #284
		wg.Add(1)

		go func() {
			for k := range masterGroups {
				resultsMutex.Lock()
				results[k] = fmt.Errorf("function panic")
				resultsMutex.Unlock()
			}

			defer wg.Done()

			for k, group := range masterGroups {
				err := group.RollingUpdate(c.Cloud, c.Force, c.MasterInterval, instanceGroups, k8sClient)

				resultsMutex.Lock()
				results[k] = err
				resultsMutex.Unlock()

				// TODO: Bail on error?
			}
		}()

		wg.Wait()
	}

	// Upgrade nodes, with greater parallelism
	{
		var wg sync.WaitGroup

		for k, nodeGroup := range nodeGroups {
			wg.Add(1)
			go func(k string, group *CloudInstanceGroup) {
				resultsMutex.Lock()
				results[k] = fmt.Errorf("function panic")
				resultsMutex.Unlock()

				defer wg.Done()

				err := group.RollingUpdate(c.Cloud, c.Force, c.NodeInterval, instanceGroups, k8sClient)

				resultsMutex.Lock()
				results[k] = err
				resultsMutex.Unlock()
			}(k, nodeGroup)
		}

		wg.Wait()
	}

	for _, err := range results {
		if err != nil {
			return err
		}
	}

	return nil
}

// CloudInstanceGroup is the AWS ASG backing an InstanceGroup
type CloudInstanceGroup struct {
	InstanceGroup *api.InstanceGroup
	ASGName       string
	Status        string
	Ready         []*CloudInstanceGroupInstance
	NeedUpdate    []*CloudInstanceGroupInstance

	asg *autoscaling.Group
}

// CloudInstanceGroupInstance describes an instance in an autoscaling group
type CloudInstanceGroupInstance struct {
	ASGInstance *autoscaling.Instance
	Node        *v1.Node
}

func (c *CloudInstanceGroup) MinSize() int {
	return int(aws.Int64Value(c.asg.MinSize))
}

func (c *CloudInstanceGroup) MaxSize() int {
	return int(aws.Int64Value(c.asg.MaxSize))
}

func buildCloudInstanceGroup(ig *api.InstanceGroup, g *autoscaling.Group, nodeMap map[string]*v1.Node) *CloudInstanceGroup {
	n := &CloudInstanceGroup{
		ASGName:       aws.StringValue(g.AutoScalingGroupName),
		InstanceGroup: ig,
		asg:           g,
	}

	readyLaunchConfigurationName := aws.StringValue(g.LaunchConfigurationName)

	for _, i := range g.Instances {
		c := &CloudInstanceGroupInstance{ASGInstance: i}

		node := nodeMap[aws.StringValue(i.InstanceId)]
		if node != nil {
			c.Node = node
		}

		if readyLaunchConfigurationName == aws.StringValue(i.LaunchConfigurationName) {
			n.Ready = append(n.Ready, c)
		} else {
			n.NeedUpdate = append(n.NeedUpdate, c)
		}
	}

	if len(n.NeedUpdate) == 0 {
		n.Status = "Ready"
	} else {
		n.Status = "NeedsUpdate"
	}

	return n
}

// Performs a rolling update on a list of nodes
func (n *CloudInstanceGroup) RollingUpdate(cloud fi.Cloud, force bool, interval time.Duration, instanceGroupList *api.InstanceGroupList, k8sClient *release_1_5.Clientset) error {
	c := cloud.(awsup.AWSCloud)

	update := n.NeedUpdate
	if force {
		update = append(update, n.Ready...)
	}
	for _, u := range update {
		instanceID := aws.StringValue(u.ASGInstance.InstanceId)
		glog.Infof("Stopping instance %q in AWS ASG %q", instanceID, n.ASGName)

		// TODO: Drain through k8s first?
		// TODO: Drain is not working well with PetSet so we will just cordon right now
		kubeCtl := &Kubectl{}
		_, err := kubeCtl.Cordon(u.Node.Name)

		if err != nil {
			// what do we do here??
			glog.V(2).Infof("Cordon failed %s - %q in AWS ASG %q", u.Node.Name, instanceID, n.ASGName)
		}

		// TODO: Temporarily increase size of ASG?
		// TODO: Remove from ASG first so status is immediately updated?
		// TODO: Batch termination, like a rolling-update

		request := &ec2.TerminateInstancesInput{
			InstanceIds: []*string{u.ASGInstance.InstanceId},
		}
		_, err = c.EC2().TerminateInstances(request)
		if err != nil {
			return fmt.Errorf("error deleting instance %q: %v", instanceID, err)
		}

		// Wait for new EC2 instances to be created
		time.Sleep(interval)

		// Wait until the cluster is happy
		// TODO: do we need to respect cloud only??
		for i := 0; i < retries; i++ {

			_, validateDidNotPass := validate.ValidateCluster(u.Node.ClusterName, instanceGroupList, k8sClient)

			if validateDidNotPass != nil {
				glog.V(2).Infof("Unable to validate kuberetes cluster %s, %v.", u.Node.ClusterName, validateDidNotPass)
				time.Sleep(interval)
			} else {
				glog.V(2).Infof("Cluster %s is validated and proceeding.", u.Node.ClusterName)
				break
			}
		}
	}

	return nil
}

func (g *CloudInstanceGroup) Delete(cloud fi.Cloud) error {
	c := cloud.(awsup.AWSCloud)

	// TODO: Graceful?

	// Delete ASG
	{
		asgName := aws.StringValue(g.asg.AutoScalingGroupName)
		glog.V(2).Infof("Deleting autoscaling group %q", asgName)
		request := &autoscaling.DeleteAutoScalingGroupInput{
			AutoScalingGroupName: g.asg.AutoScalingGroupName,
			ForceDelete:          aws.Bool(true),
		}
		_, err := c.Autoscaling().DeleteAutoScalingGroup(request)
		if err != nil {
			return fmt.Errorf("error deleting autoscaling group %q: %v", asgName, err)
		}
	}

	// Delete LaunchConfig
	{
		lcName := aws.StringValue(g.asg.LaunchConfigurationName)
		glog.V(2).Infof("Deleting autoscaling launch configuration %q", lcName)
		request := &autoscaling.DeleteLaunchConfigurationInput{
			LaunchConfigurationName: g.asg.LaunchConfigurationName,
		}
		_, err := c.Autoscaling().DeleteLaunchConfiguration(request)
		if err != nil {
			return fmt.Errorf("error deleting autoscaling launch configuration %q: %v", lcName, err)
		}
	}

	return nil
}

func (n *CloudInstanceGroup) String() string {
	return "CloudInstanceGroup:" + n.ASGName
}
