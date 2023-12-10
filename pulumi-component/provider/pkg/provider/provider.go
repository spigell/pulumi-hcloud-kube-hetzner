// Copyright 2021, Pulumi Corporation.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package provider

import (
	"github.com/pulumi/pulumi/pkg/v3/resource/provider"
	"github.com/pulumi/pulumi/sdk/v3/go/common/util/cmdutil"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	pp "github.com/pulumi/pulumi/sdk/v3/go/pulumi/provider"
)

const (
	ProviderName  = "hcloud-kube-hetzner"
	ComponentName = ProviderName + ":index:Cluster"
)

// Serve launches the gRPC server for the resource provider.
func Serve(version string, schema []byte) {
	if err := provider.ComponentMain(ProviderName, version, schema, Construct); err != nil {
		cmdutil.ExitError(err.Error())
	}
}

// Construct is the RPC call that initiates the creation of a new component resource. It
// creates, registers, and returns the resulting object.
func Construct(ctx *pulumi.Context, typ, name string, inputs pp.ConstructInputs,
	opts pulumi.ResourceOption) (*pp.ConstructResult, error) {
	return construct(ctx, &Cluster{}, typ, name, &ClusterArgs{}, inputs, opts)
}
