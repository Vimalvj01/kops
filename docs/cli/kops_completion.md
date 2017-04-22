## kops completion

Output shell completion code for the given shell (bash or zsh).

### Synopsis




Output shell completion code for the specified shell (bash or zsh).
The shell code must be evalutated to provide interactive
completion of kops commands.  This can be done by sourcing it from
the .bash_profile.

Note: this requires the bash-completion framework, which is not installed
by default on Mac.  This can be installed by using homebrew:

	$ brew install bash-completion

Once installed, bash_completion must be evaluated.  This can be done by adding the
following line to the .bash_profile

	$ source $(brew --prefix)/etc/bash_completion

Note for zsh users: [1] zsh completions are only supported in versions of zsh >= 5.2

```
kops completion
```

### Examples

```


# Install bash completion on a Mac using homebrew

$ brew install bash-completion

$ printf "
# Bash completion support
source $(brew --prefix)/etc/bash_completion
" >> $HOME/.bash_profile
	source $HOME/.bash_profile
	# Load the kops completion code for bash into the current shell
	source <(kops completion bash)
	# Write bash completion code to a file and source if from .bash_profile
	kops completion bash > ~/.kops/completion.bash.inc
	printf "

# kops shell completion
'$HOME/.kops/completion.bash.inc'
" >> $HOME/.bash_profile

$ source $HOME/.bash_profile

# Load the kops completion code for zsh[1] into the current shell
source <(kops completion zsh)
```

### Options

```
      --shell string   target shell (bash).
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

