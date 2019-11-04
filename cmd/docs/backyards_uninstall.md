## backyards uninstall

Uninstall Backyards

### Synopsis

Uninstall Backyards

The command automatically removes the resources.
It can only dump the removable resources with the '--dump-resources' option.

```
backyards uninstall [flags]
```

### Examples

```
  # Default uninstall
  backyards uninstall

  # Uninstall Backyards from a non-default namespace
  backyards uninstall -n backyards-system
```

### Options

```
  -d, --dump-resources           Dump resources to stdout instead of applying them
  -h, --help                     help for uninstall
      --istio-namespace string   Namespace of Istio sidecar injector (default "istio-system")
      --release-name string      Name of the release (default "backyards")
  -a, --uninstall-everything     Uninstall all components at once
```

### Options inherited from parent commands

```
      --accept-license                  Accept the license: https://banzaicloud.com/docs/backyards/evaluation-license
      --base-url string                 Custom Backyards base URL (uses port forwarding or proxying if empty)
      --cacert string                   The CA to use for verifying Backyards' server certificate
      --color                           use colors on non-tty outputs (default true)
      --context string                  name of the kubeconfig context to use
      --formatting.force-color          force color even when non in a terminal
      --interactive                     ask questions interactively even if stdin or stdout is non-tty
  -c, --kubeconfig string               path to the kubeconfig file to use for CLI requests
  -p, --local-port int                  Use this local port for port forwarding / proxying to Backyards (when set to 0, a random port will be used) (default -1)
  -n, --namespace string                Namespace in which Backyards is installed [$BACKYARDS_NAMESPACE] (default "backyards-system")
      --non-interactive                 never ask questions interactively
  -o, --output string                   output format (table|yaml|json) (default "table")
      --persistent-config-file string   Backyards persistent config file to use instead of the default at ~/.banzai/backyards/
      --token string                    Authentication token to use to communicate with Backyards
      --use-portforward                 Use port forwarding instead of proxying to reach Backyards
  -v, --verbose                         turn on debug logging
```

### SEE ALSO

* [backyards](backyards.md)	 - Install and manage Backyards

