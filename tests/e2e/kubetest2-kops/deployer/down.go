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
	"strings"

	"k8s.io/klog/v2"
	"sigs.k8s.io/kubetest2/pkg/exec"
)

func (d *deployer) Down() error {
	if err := d.init(); err != nil {
		return err
	}
	if err := d.DumpClusterLogs(); err != nil {
		klog.Warningf("Dumping cluster logs at the start of Down() failed: %s", err)
	}

	args := []string{
		d.KopsBinaryPath, "delete", "cluster",
		"--name", d.ClusterName,
		"--yes",
	}
	klog.Info(strings.Join(args, " "))
	cmd := exec.Command(args[0], args[1:]...)
	cmd.SetEnv(d.env()...)

	exec.InheritOutput(cmd)
	return cmd.Run()
}
