/*
Copyright 2020 The Kubernetes Authors.

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

package deployer

import (
	"fmt"
	"os"
	osexec "os/exec"
	"strings"

	"k8s.io/klog/v2"
	"k8s.io/kops/tests/e2e/kubetest2-kops/aws"
	"k8s.io/kops/tests/e2e/kubetest2-kops/do"
	"k8s.io/kops/tests/e2e/kubetest2-kops/gce"
	"k8s.io/kops/tests/e2e/kubetest2-kops/util"
	"sigs.k8s.io/kubetest2/pkg/exec"
)

func (d *deployer) Up() error {
	if err := d.init(); err != nil {
		return err
	}

	publicIP, err := util.ExternalIPRange()
	if err != nil {
		return err
	}

	adminAccess := d.AdminAccess
	if adminAccess == "" {
		adminAccess = publicIP
	}

	zones, err := d.zones()
	if err != nil {
		return err
	}

	if d.TemplatePath != "" {
		values, err := d.templateValues(zones, adminAccess)
		if err != nil {
			return err
		}
		if err := d.renderTemplate(values); err != nil {
			return err
		}
		if err := d.replace(); err != nil {
			return err
		}
	} else {
		err := d.createCluster(zones, adminAccess)
		if err != nil {
			return err
		}
	}
	isUp, err := d.IsUp()
	if err != nil {
		return err
	} else if isUp {
		klog.V(1).Infof("cluster reported as up")
	} else {
		klog.Errorf("cluster reported as down")
	}
	return nil
}

func (d *deployer) createCluster(zones []string, adminAccess string) error {

	args := []string{
		d.KopsBinaryPath, "create", "cluster",
		"--name", d.ClusterName,
		"--admin-access", adminAccess,
		"--cloud", d.CloudProvider,
		"--kubernetes-version", d.KubernetesVersion,
		"--master-count", "1",
		"--master-volume-size", "48",
		"--node-count", "4",
		"--node-volume-size", "48",
		"--override", "cluster.spec.nodePortAccess=0.0.0.0/0",
		"--ssh-public-key", d.SSHPublicKeyPath,
		"--yes",
	}

	if d.CloudProvider == "aws" {
		zones, err := aws.RandomZones(1)
		if err != nil {
			return err
		}
		args = append(args, "--zones", strings.Join(zones, ","))

		args = append(args, "--master-size", "c5.large")
	}

	if d.CloudProvider == "gce" {
		zones, err := gce.RandomZones(1)
		if err != nil {
			return err
		}
		args = append(args, "--zones", strings.Join(zones, ","))

		args = append(args, "--master-size", "e2-standard-2")
	}

	if d.CloudProvider == "digitalocean" {
		zones, err := do.RandomZones(1)
		if err != nil {
			return err
		}
		args = append(args, "--zones", strings.Join(zones, ","))
		args = append(args, "--master-size", "s-8vcpu-16gb")
	}

	if d.Networking != "" {
		args = append(args, "--networking", d.Networking)
	}

	klog.Info(strings.Join(args, " "))
	cmd := exec.Command(args[0], args[1:]...)
	cmd.SetEnv(d.env()...)

	exec.InheritOutput(cmd)
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

func (d *deployer) IsUp() (bool, error) {
	args := []string{
		d.KopsBinaryPath, "validate", "cluster",
		"--name", d.ClusterName,
		"--wait", "15m",
	}
	klog.Info(strings.Join(args, " "))

	cmd := exec.Command(args[0], args[1:]...)
	cmd.SetEnv(d.env()...)

	exec.InheritOutput(cmd)
	err := cmd.Run()
	// `kops validate cluster` exits 2 if validation failed
	if exitErr, ok := err.(*osexec.ExitError); ok && exitErr.ExitCode() == 2 {
		return false, nil
	}
	return err == nil, err
}

// verifyUpFlags ensures fields are set for creation of the cluster
func (d *deployer) verifyUpFlags() error {
	// These environment variables are defined by the "preset-aws-ssh" prow preset
	// https://github.com/kubernetes/test-infra/blob/3d3b325c98b739b526ba5d93ce21c90a05e1f46d/config/prow/config.yaml#L653-L670
	if d.SSHPrivateKeyPath == "" {
		d.SSHPrivateKeyPath = os.Getenv("AWS_SSH_PRIVATE_KEY_FILE")
	}
	if d.SSHPublicKeyPath == "" {
		d.SSHPublicKeyPath = os.Getenv("AWS_SSH_PUBLIC_KEY_FILE")
	}

	return nil
}

func (d *deployer) zones() ([]string, error) {
	switch d.CloudProvider {
	case "aws":
		return aws.RandomZones(1)
	case "gce":
		return gce.RandomZones(1)
	}
	return nil, fmt.Errorf("unsupported CloudProvider: %v", d.CloudProvider)
}
