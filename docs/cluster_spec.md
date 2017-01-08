# Description of Keys in `config` and `cluster.spec`

This list is not complete, but aims to document any keys that are less than self-explanatory.

## spec

### adminAccess

This array configures the CIDRs that are able to ssh into nodes. On AWS this is manifested as inbound security group rules on the `nodes` and `master` security groups.

Use this key to restrict cluster access to an office ip address range, for example.

```yaml
spec:
  adminAccess:
    - 12.34.56.78/32
```

### cluster.spec Subnet Keys

#### subnetId
ID of a subnet to share in an existing VPC.

#### ngwId/ngwEip
NgwId: ID of an existing AWS NAT Gateway (NGW) to be used for a Private subnet.
NgwEip: ID of the AWS ElasticIP allocation connected to the specified NGW

If you wish to use a shared NGW, you MUST specify the ElasticIP associated with it. At this time, there is no reason to specify an ElasticIP without a corresponding NGW.

```yaml
spec:
  subnets:
  - cidr: 10.20.64.0/21
    name: us-east-1a
    ngwEip: eipalloc-12345
    ngwId: nat-987654321
    type: Private
    zone: us-east-1a
  - cidr: 10.20.32.0/21
    name: utility-us-east-1a
    subnetId: subnet-12345
    type: Utility
    zone: us-east-1a
```

### kubeAPIServer

This block contains configuration for the `kube-apiserver`.

#### runtimeConfig

Keys and values here are translated into `--runtime-config` values for `kube-apiserver`, separated by commas.

Use this to enable alpha features, for example:

```yaml
spec:
  kubeAPIServer:
    runtimeConfig:
      batch/v2alpha1: "true"
      apps/v1alpha1: "true"
```

Will result in the flag `--runtime-config=batch/v2alpha1=true,apps/v1alpha1=true`. Note that `kube-apiserver` accepts `true` as a value for switch-like flags.

### networkID

On AWS, this is the id of the VPC the cluster is created in. If creating a cluster from scratch, this field doesn't need to be specified at create time; `kops` will create a `VPC` for you.

```yaml
spec:
  networkID: vpc-abcdefg1
```
