## kops update

Creates or updates cloud resources to match cluster spec.

### Synopsis


Update clusters.

### Examples

```
  # After cluster has been created, configure it with:
  kops update cluster k8s.cluster.site --yes --state=s3://kops-state-1234
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
* [kops](kops.md)	 - kops is kubernetes ops
* [kops update cluster](kops_update_cluster.md)	 - Create or update cloud or cluster resources to match current cluster state.
* [kops update federation](kops_update_federation.md)	 - Update federation

