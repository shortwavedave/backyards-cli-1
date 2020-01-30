## backyards tap

Tap into HTTP/GRPC mesh traffic

### Synopsis

Tap into HTTP/GRPC mesh traffic

```
backyards tap [[ns|workload|pod]/resource-name] [flags]
```

### Examples

```

  # tap the movies-v1 deployment in the default namespace
  backyards tap workload/movies-v1

  # tap the movies-v1-7f9645bfd7-8vgm9 pod in the default namespace
  backyards tap pod/movies-v1-7f9645bfd7-8vgm9

  # tap the backyards-demo namespace
  backyards tap ns/backyards-demo

  # tap the backyards-demo namespace request to test namespace
  backyards tap ns/backyards-demo --destination ns/test
```

### Options

```
      --authority string        Show requests with this authority
      --destination string      Show requests to this resource
      --destination-ns string   Namespace of the destination resource; by default the current "--namespace" is used
      --direction string        Show requests with this direction (inbound|outbound)
  -h, --help                    help for tap
      --method string           Show requests with this request method
      --ns string               Namespace of the specified resource (default "default")
      --path string             Show requests with paths with this prefix
      --response-code uints     Show request with this response code (default [])
      --scheme string           Show requests with this scheme
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

* [backyards](backyards.md)	 - Install and manage Backyards

