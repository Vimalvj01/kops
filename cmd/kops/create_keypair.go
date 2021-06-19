/*
Copyright 2021 The Kubernetes Authors.

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

package main

import (
	"io"

	"github.com/spf13/cobra"
	"k8s.io/kops/cmd/kops/util"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
)

var (
	createKeypairLong = templates.LongDesc(i18n.T(`
	Create a keypair`))

	createKeypairExample = templates.Examples(i18n.T(`
	Add a cluster CA certificate and private key.
	kops create keypair ca \
		--cert ~/ca.pem --key ~/ca-key.pem \
		--name k8s-cluster.example.com --state s3://my-state-store
	`))

	createKeypairShort = i18n.T(`Create a keypair.`)
)

func NewCmdCreateKeypair(f *util.Factory, out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "keypair",
		Short:   createKeypairShort,
		Long:    createKeypairLong,
		Example: createKeypairExample,
	}

	// create subcommands
	cmd.AddCommand(NewCmdCreateKeypairCa(f, out))

	return cmd
}
