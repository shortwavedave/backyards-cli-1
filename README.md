<p align="center"><img src="docs/img/backyards-logo.svg" width="260"></p>

This is a command line interface for [Backyards](https://banzaicloud.com/blog/backyards-release-1-0/), the Banzai Cloud automated service mesh, built on Istio.

## TL;DR

To install and launch Backyards with a demo application in a few minutes, run the following command (`KUBECONFIG` must be set for your cluster):

```bash
curl https://getbackyards.sh | sh && backyards install -a --run-demo
```

## Installation

Run the following command:

```bash
curl https://getbackyards.sh
```

Alternatively, pre-built binaries are available in multiple package formats. Download the [latest release](https://github.com/banzaicloud/backyards-cli/releases).

## Build from source

To build a binary (under `build/`) from the source code, clone the repo and then run from the root:

```bash
$ make build
```

## Usage

### Quick start

To install Istio, all Backyards components and a demo application on a brand new cluster, you just need to issue one command (`KUBECONFIG` must be set for your cluster):

```bash
$ backyards install -a --run-demo
```

This command first installs Istio with our open-source [Istio operator](https://github.com/banzaicloud/istio-operator), then installs Backyards components as well as a demo application for demonstration purposes. After the installation of each component has finished, the Backyards UI will automatically open and send some traffic to the demo application. **By issuing this one simple command you can watch as Backyards starts a brand new Istio service mesh in just a few minutes!**

### Install/Uninstall components

The following components can be installed/uninstalled individually as well with the CLI (the `-a` flag installs/uninstalls them all):

- [istio](cmd/docs/backyards_istio.md): `backyards istio [install|uninstall]`
- [canary-operator](cmd/docs/backyards_canary.md): `backyards canary [install|uninstall]`
- [cert-manager](cmd/docs/backyards_cert-manager.md): `backyards cert-manager [install|uninstall]`
- [backyards (backend and UI)](cmd/docs/backyards.md): `backyards [install|uninstall]`
- [demo application](cmd/docs/backyards_demoapp.md): `backyards demoapp [install|uninstall]`

### Handy features

- Istio can be installed with a customized CR with: `backyards istio install -f your_istio_cr.yaml`
- The Backyards UI can be opened with: `backyards dashboard`
- You can display a graph with the most important RED metrics of your cluster with: `backyards graph`
- [Traffic Shifting](docs/traffic_shifting.md) can be configured
- [Circuit Breaking](docs/circuit_breaking.md) can be configured

### All commands

```text
Install and manage Backyards

Usage:
  backyards [flags]


Commands:
  install      Install Backyards
  uninstall    Uninstall Backyards

  dashboard    Open the Backyards dashboard in a web browser
  graph        Show graph
  login        Log in to Backyards (aliases: l)
  routing      Manage service routing configurations (aliases: r)
  sidecar-proxy Manage sidecar-proxy related configurations (aliases: sp)

  config        View and manage persistent configuration
  version      Print the client and api version information
  help         Help about any command


Components commands:
  canary       Install and manage Canary feature
  cert-manager Install and manage cert-manager
  demoapp      Install and manage demo application
  istio        Install and manage Istio

Flags:
  -u, --base-url string     Custom Backyards base URL. Uses automatic port forwarding / proxying if empty
      --cacert string       The CA to use for verifying Backyards' server certificate
      --context string      name of the kubeconfig context to use
  -h, --help                help for backyards
      --interactive         ask questions interactively even if stdin or stdout is non-tty
  -c, --kubeconfig string   path to the kubeconfig file to use for CLI requests
  -p, --local-port int      Use this local port for port forwarding / proxying to Backyards (when set to 0, a random port will be used) (default -1)
  -n, --namespace string    namespace in which Backyards is installed [$BACKYARDS_NAMESPACE] (default "backyards-system")
      --non-interactive     never ask questions interactively
  -o, --output string       output format (table|yaml|json) (default "table")
      --use-portforward     Use port forwarding instead of proxying to reach Backyards
  -v, --verbose             turn on debug logging
      --version             version for backyards

Use "backyards [command] --help" for more information about a command.
```

### Cleanup

To remove the demo application, Backyards, and Istio from your cluster, you just need to apply one command, which takes care of removing these components in the correct order:

```bash
$ backyards uninstall -a
```
