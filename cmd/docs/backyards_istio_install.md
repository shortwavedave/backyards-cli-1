## backyards istio install

Installs Istio utilizing Banzai Cloud's Istio-operator

### Synopsis

Installs Istio utilizing Banzai Cloud's Istio-operator.

The command automatically applies the resources.
It can only dump the applicable resources with the '--dump-resources' option.

The manual mode is a two phase process as the operator needs custom CRDs to work.
The installer automatically detects whether the CRDs are installed or not, and behaves accordingly.

```
backyards istio install [flags]
```

### Examples

```
  # Default install.
  backyards istio install

  # Install Istio into a non-default namespace.
  backyards istio install -n istio-custom-ns
```

### Options

```
  -d, --dump-resources         Dump resources to stdout instead of applying them
      --force                  Force Istio upgrade (only applicable in non-interactive mode)
  -h, --help                   help for install
  -f, --istio-cr-file string   Filename of a custom Istio CR yaml
      --release-name string    Name of the release (default "istio-operator")
      --without-istiod         Use multi-component install without Istiod
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

