## backyards routing traffic-shifting

Manage traffic-shifting configurations

### Synopsis

Manage traffic-shifting configurations

### Options

```
  -h, --help   help for traffic-shifting
```

### Options inherited from parent commands

```
  -u, --base-url string     Custom Backyards base URL. Uses automatic port forwarding / proxying if empty
      --cacert string       The CA to use for verifying Backyards' server certificate
      --context string      name of the kubeconfig context to use
      --interactive         ask questions interactively even if stdin or stdout is non-tty
  -c, --kubeconfig string   path to the kubeconfig file to use for CLI requests
  -p, --local-port int      Use this local port for port forwarding / proxying to Backyards (when set to 0, a random port will be used) (default -1)
  -n, --namespace string    namespace in which Backyards is installed [$BACKYARDS_NAMESPACE] (default "backyards-system")
      --non-interactive     never ask questions interactively
  -o, --output string       output format (table|yaml|json) (default "table")
      --use-portforward     Use port forwarding instead of proxying to reach Backyards
  -v, --verbose             turn on debug logging
```

### SEE ALSO

* [backyards routing](backyards_routing.md)	 - Manage service routing configurations
* [backyards routing traffic-shifting delete](backyards_routing_traffic-shifting_delete.md)	 - Delete traffic shifting rules of a service
* [backyards routing traffic-shifting get](backyards_routing_traffic-shifting_get.md)	 - Get traffic shifting rules for a service
* [backyards routing traffic-shifting set](backyards_routing_traffic-shifting_set.md)	 - Set traffic shifting rules for a service

