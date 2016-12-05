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

package test

import (
	"testing"
	"time"
	"fmt"
)

const (
	// Hang for 3 minutes waiting for the API to come up
	// It usually takes ~about~ 2 minutes - so we bake in
	// an extra 60 seconds for good measure
	ApiTimeoutIterations = 300
	ApiTimeoutDuration = time.Second * 1
)

const KOPS_VALIDATE_CLUSTER = `validate cluster --name %s --state % -v %c`

func TestValidate(t *testing.T) {
	err := Validate()
	if err != nil {
		t.Error(err)
	}
}

func Validate() error {
	kopsValidationCommand := fmt.Sprintf(KOPS_VALIDATE_CLUSTER, TestClusterName, TestStateStore, TestVerbosity)
	for i := 0; i <= ApiTimeoutIterations; i++ {
		_ , stderr := ExecOuput(KopsPath, kopsValidationCommand, []string{})
		if stderr != nil {
			if i == ApiTimeoutIterations {
				return fmt.Errorf("Unable to validate after timeout: %v", stderr)
			}
			time.Sleep(ApiTimeoutDuration)
			break
		}
	}
	return nil
}
