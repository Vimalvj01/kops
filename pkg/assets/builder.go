/*
Copyright 2017 The Kubernetes Authors.

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

package assets

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"k8s.io/kops/pkg/featureflag"
	"k8s.io/kops/pkg/kubemanifest"
	"k8s.io/kops/pkg/tasks"
)

// RewriteManifests controls whether we rewrite manifests
// Because manifest rewriting converts everything to and from YAML, we normalize everything by doing so
var RewriteManifests = featureflag.New("RewriteManifests", featureflag.Bool(true))

// AssetBuilder discovers and remaps assets
type AssetBuilder struct {
	Assets []*Asset
}

type Asset struct {
	Origin string
	Mirror string
}

func NewAssetBuilder() *AssetBuilder {
	return &AssetBuilder{}
}

// RemapManifest transforms a kubernetes manifest.
// Whenever we are building a Task that includes a manifest, we should pass it through RemapManifest first.
// This will:
// * rewrite the images if they are being redirected to a mirror, and ensure the image is uploaded
func (a *AssetBuilder) RemapManifest(tasks tasks.TaskBuilderContext, data []byte) ([]byte, error) {
	if !RewriteManifests.Enabled() {
		return data, nil
	}
	manifests, err := kubemanifest.LoadManifestsFrom(data)
	if err != nil {
		return nil, err
	}

	var yamlSeparator = []byte("\n---\n\n")
	var remappedManifests [][]byte
	for _, manifest := range manifests {
		err := manifest.RemapImages(func(image string) (string, error) {
			return a.remapImage(tasks, image)
		})
		if err != nil {
			return nil, fmt.Errorf("error remapping images: %v", err)
		}
		y, err := manifest.ToYAML()
		if err != nil {
			return nil, fmt.Errorf("error re-marshalling manifest: %v", err)
		}

		remappedManifests = append(remappedManifests, y)
	}

	return bytes.Join(remappedManifests, yamlSeparator), nil
}

func (a *AssetBuilder) remapImage(tasks tasks.TaskBuilderContext, image string) (string, error) {
	asset := &Asset{}

	asset.Origin = image

	if strings.HasPrefix(image, "kope/dns-controller:") {
		// To use user-defined DNS Controller:
		// 1. DOCKER_REGISTRY=[your docker hub repo] make dns-controller-push
		// 2. export DNSCONTROLLER_IMAGE=[your docker hub repo]
		// 3. make kops and create/apply cluster
		override := os.Getenv("DNSCONTROLLER_IMAGE")
		if override != "" {
			image = override
		}
	}

	asset.Mirror = image

	if asset.Origin != asset.Mirror {
		// TODO: Add task to copy docker image
	}

	a.Assets = append(a.Assets, asset)

	return image, nil
}
