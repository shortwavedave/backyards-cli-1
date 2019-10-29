## backyards routing circuit-breaker delete

Delete circuit breaker rules of a service

### Synopsis

Delete circuit breaker rules of a service

```
backyards routing circuit-breaker delete [[--service=]namespace/servicename] [flags]
```

### Options

```
  -h, --help             help for delete
      --service string   Service name
```

### Options inherited from parent commands

```
      --base-url string                 Custom Backyards base URL (uses port forwarding or proxying if empty)
      --cacert string                   The CA to use for verifying Backyards' server certificate
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

* [backyards routing circuit-breaker](backyards_routing_circuit-breaker.md)	 - Manage circuit-breaker configurations

