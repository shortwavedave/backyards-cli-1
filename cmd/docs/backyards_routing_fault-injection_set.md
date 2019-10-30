## backyards routing fault-injection set

Set fault injection for http route of a service

### Synopsis

Set fault injection for http route of a service

```
backyards routing fault-injection set [[--service=]namespace/servicename] [[--match=]field:kind=value] ... [flags]
```

### Options

```
      --abort-percentage float32     Percentage of requests on which the abort will be injected
      --abort-status-code int        HTTP status code to use to abort the HTTP request
      --delay-fixed-delay duration   Add a fixed delay before forwarding the request. Format: 1h/1m/1s/1ms. MUST be >=1ms.
      --delay-percentage float32     Percentage of requests on which the delay will be injected
  -h, --help                         help for set
  -m, --match stringArray            HTTP request match
      --service string               Service name
```

### Options inherited from parent commands

```
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

* [backyards routing fault-injection](backyards_routing_fault-injection.md)	 - Manage fault injection configurations

