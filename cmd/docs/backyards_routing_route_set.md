## backyards routing route set

Set http route for a service

### Synopsis

Set http route for a service

```
backyards routing route set [[--service=]namespace/servicename] [[--match=]field:kind=value] ... [flags]
```

### Options

```
  -h, --help                             help for set
  -m, --match stringArray                HTTP request match
      --redirect-authority string        overwrite the Authority/Host portion of the URL with this value
      --redirect-uri string              overwrite the Path portion of the URL with this value
      --retry-attempts int               Number of retries for a given request (default -1)
      --retry-on string                  Specifies the conditions under which retry takes place
      --retry-per-try-timeout duration   Timeout per retry attempt for a given request. format: 1h/1m/1s/1ms, must be >=1ms (default 2s)
  -d, --route-destination stringArray    HTTP route destination
  -w, --route-weights string             The proportions of traffic to be forwarded to the route destinations. (0-100) (default "100")
      --service string                   Service name
  -t, --timeout duration                 Timeout for HTTP requests (default -1ns)
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

* [backyards routing route](backyards_routing_route.md)	 - Manage route configurations

