## backyards sidecar-proxy auto-inject

Manage auto-injection configurations

### Synopsis

Manage auto-injection configurations

### Options

```
  -h, --help   help for auto-inject
```

### Options inherited from parent commands

```
      --accept-license                  Accept the license: https://banzaicloud.com/docs/backyards/evaluation-license
      --backyards-namespace string      Namespace in which Backyards is installed [$BACKYARDS_NAMESPACE] (default "backyards-system")
      --base-url string                 Custom Backyards base URL (uses port forwarding or proxying if empty)
      --cacert string                   The CA to use for verifying Backyards' server certificate
      --color                           use colors on non-tty outputs (default true)
      --context string                  name of the kubeconfig context to use
      --formatting.force-color          force color even when non in a terminal
      --interactive                     ask questions interactively even if stdin or stdout is non-tty
  -c, --kubeconfig string               path to the kubeconfig file to use for CLI requests
  -p, --local-port int                  Use this local port for port forwarding / proxying to Backyards (when set to 0, a random port will be used) (default -1)
      --non-interactive                 never ask questions interactively
  -o, --output string                   output format (table|yaml|json) (default "table")
      --persistent-config-file string   Backyards persistent config file to use instead of the default at ~/.banzai/backyards/
      --token string                    Authentication token to use to communicate with Backyards
      --use-portforward                 Use port forwarding instead of proxying to reach Backyards
  -v, --verbose                         turn on debug logging
```

### SEE ALSO

* [backyards sidecar-proxy](backyards_sidecar-proxy.md)	 - Manage sidecar-proxy related configurations
* [backyards sidecar-proxy auto-inject off](backyards_sidecar-proxy_auto-inject_off.md)	 - Disable sidecar injection for the given namespace
* [backyards sidecar-proxy auto-inject on](backyards_sidecar-proxy_auto-inject_on.md)	 - Enable sidecar injection for the given namespace

