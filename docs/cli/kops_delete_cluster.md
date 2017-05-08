## kops delete cluster

Delete a cluster.

### Synopsis


Deletes a Kubneretes cluster and all associated resources.  Resources include instancegroups, and the state store.  There is no "UNDO" for this command.

```
kops delete cluster CLUSTERNAME [--yes]
```

### Examples

```
  # Delete a cluster.
  kops delete cluster --name=k8s.cluster.site --yes
  
  # Delete an instancegroup for the k8s-cluster.example.com cluster.
  # The --yes option runs the command immediately.
  kops delete ig --name=k8s-cluster.example.com node-example --yes
```

### Options

```
      --external        Delete an external cluster
      --region string   region
      --unregister      Don't delete cloud resources, just unregister the cluster
  -y, --yes             Specify --yes to delete the cluster
```

### Options inherited from parent commands

```
      --alsologtostderr                  log to standard error as well as files
      --config string                    config file (default is $HOME/.kops.yaml)
      --log_backtrace_at traceLocation   when logging hits line file:N, emit a stack trace (default :0)
      --log_dir string                   If non-empty, write log files in this directory
      --logtostderr                      log to standard error instead of files (default false)
      --name string                      Name of cluster
      --state string                     Location of state storage
      --stderrthreshold severity         logs at or above this threshold go to stderr (default 2)
  -v, --v Level                          log level for V logs
      --vmodule moduleSpec               comma-separated list of pattern=N settings for file-filtered logging
```

### SEE ALSO
* [kops delete](kops_delete.md)	 - Delete clusters,instancegroups, or secrets.

