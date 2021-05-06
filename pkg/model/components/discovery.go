/*
Copyright 2019 The Kubernetes Authors.

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

package components

import (
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/model/iam"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/loader"
)

// DiscoveryOptionsBuilder adds options for identity discovery to the model (mostly kube-apiserver)
type DiscoveryOptionsBuilder struct {
	*OptionsContext
}

var _ loader.OptionsBuilder = &DiscoveryOptionsBuilder{}

func (b *DiscoveryOptionsBuilder) BuildOptions(o interface{}) error {
	clusterSpec := o.(*kops.ClusterSpec)

	if b.IsKubernetesLT("1.20") {
		if clusterSpec.KubeAPIServer.FeatureGates == nil {
			return nil
		}
		if _, ok := clusterSpec.KubeAPIServer.FeatureGates["ServiceAccountIssuerDiscovery"]; !ok {
			return nil
		}
	}

	if clusterSpec.KubeAPIServer == nil {
		clusterSpec.KubeAPIServer = &kops.KubeAPIServerConfig{}
	}

	kubeAPIServer := clusterSpec.KubeAPIServer

	if len(kubeAPIServer.APIAudiences) == 0 {
		kubeAPIServer.APIAudiences = []string{"kubernetes.svc.default"}
	}

	serviceAccountIssuer, err := iam.ServiceAccountIssuer(clusterSpec)
	if err != nil {
		return err
	}
	kubeAPIServer.ServiceAccountIssuer = &serviceAccountIssuer

	// We set apiserver ServiceAccountKey and ServiceAccountSigningKeyFile in nodeup

	if kubeAPIServer.ServiceAccountJWKSURI == nil {
		jwksURI, err := iam.ServiceAccountIssuer(clusterSpec)
		if err != nil {
			return err
		}
		kubeAPIServer.ServiceAccountJWKSURI = fi.String(jwksURI + "/openid/v1/jwks")
	}

	return nil
}
