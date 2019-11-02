## backyards canary uninstall

Output or delete Kubernetes resources to uninstall canary feature

### Synopsis

Output or delete Kubernetes resources to uninstall canary feature.

The command automatically removes the resources.
It can only dump the removable resources with the '--dump-resources' option.

The command can uninstall every component at once with the '--uninstall-everything' option.

```
backyards canary uninstall [flags]
```

### Examples

```
  # Default uninstall.
  backyards canary uninstall

  # Uninstall Canary feature from a non-default namespace.
  backyards canary uninstall install -n custom-istio-ns
```

### Options

```
      --canary-namespace string   Namespace for the canary operator (default "backyards-canary")
  -d, --dump-resources            Dump resources to stdout instead of applying them
  -h, --help                      help for uninstall
      --release-name string       Name of the release (default "canary-operator")
```

### Options inherited from parent commands

```
      --accept-license                  Accept the license: https://banzaicloud.com/docs/backyards/license
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

* [backyards canary](backyards_canary.md)	 - Install and manage canary feature

