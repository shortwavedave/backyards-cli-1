## backyards istio cluster

Manage Istio mesh member clusters

### Synopsis

Manage Istio mesh member clusters

### Options

```
  -h, --help   help for cluster
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
  -n, --namespace string                Namespace in which Istio is installed [$ISTIO_NAMESPACE] (default "istio-system")
      --non-interactive                 never ask questions interactively
  -o, --output string                   output format (table|yaml|json) (default "table")
      --persistent-config-file string   Backyards persistent config file to use instead of the default at ~/.banzai/backyards/
      --token string                    Authentication token to use to communicate with Backyards
      --use-portforward                 Use port forwarding instead of proxying to reach Backyards
  -v, --verbose                         turn on debug logging
```

### SEE ALSO

* [backyards istio](backyards_istio.md)	 - Install and manage Istio
* [backyards istio cluster attach](backyards_istio_cluster_attach.md)	 - Attach peer cluster to the mesh
* [backyards istio cluster detach](backyards_istio_cluster_detach.md)	 - Detach peer cluster from the mesh
* [backyards istio cluster status](backyards_istio_cluster_status.md)	 - Show cluster status

