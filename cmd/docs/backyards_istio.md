## backyards istio

Install and manage Istio

### Synopsis

Install and manage Istio

### Options

```
  -h, --help   help for istio
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

* [backyards](backyards.md)	 - Install and manage Backyards
* [backyards istio install](backyards_istio_install.md)	 - Installs Istio utilizing Banzai Cloud's Istio-operator
* [backyards istio uninstall](backyards_istio_uninstall.md)	 - Output or delete Kubernetes resources to uninstall Istio

