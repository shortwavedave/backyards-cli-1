## backyards

Install and manage Backyards

### Synopsis

Install and manage Backyards

### Options

```
      --accept-license                  Accept the license: https://banzaicloud.com/docs/backyards/license
      --base-url string                 Custom Backyards base URL (uses port forwarding or proxying if empty)
      --cacert string                   The CA to use for verifying Backyards' server certificate
      --color                           use colors on non-tty outputs (default true)
      --context string                  name of the kubeconfig context to use
      --formatting.force-color          force color even when non in a terminal
  -h, --help                            help for backyards
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
* [backyards cert-manager](backyards_cert-manager.md)	 - Install and manage cert-manager
* [backyards config](backyards_config.md)	 - View and manage persistent configuration
* [backyards dashboard](backyards_dashboard.md)	 - Open the Backyards dashboard in a web browser
* [backyards demoapp](backyards_demoapp.md)	 - Install and manage demo application
* [backyards graph](backyards_graph.md)	 - Show graph
* [backyards install](backyards_install.md)	 - Install Backyards
* [backyards istio](backyards_istio.md)	 - Install and manage Istio
* [backyards license](backyards_license.md)	 - Shows Backyards license
* [backyards login](backyards_login.md)	 - Log in to Backyards
* [backyards routing](backyards_routing.md)	 - Manage service routing configurations
* [backyards sidecar-proxy](backyards_sidecar-proxy.md)	 - Manage sidecar-proxy related configurations
* [backyards uninstall](backyards_uninstall.md)	 - Uninstall Backyards
* [backyards version](backyards_version.md)	 - Print the client and api version information

