# kubectl-view-gather

A `kubectl` plugin to **browse and query Kubernetes _must-gather_ dumps** using familiar `kubectl` commands.

## ✨ Overview

`kubectl-view-gather` enables interactive exploration of Kubernetes _must-gather_ archives as if you were querying a live cluster.
This simplifies diagnostics and post-mortem analysis by providing a `kubectl`-like interface to examine gathered state.

Must-gather dumps contain snapshots of Kubernetes resources (e.g., Pods, Deployments, Events) in raw YAML format. 
This plugin lets you query them like:

```sh
kubectl view gather --must-gather-path=~/my-must-gather get pods --namespace my-namespace
kubectl view gather --must-gather-path=~/my-must-gather describe node my-node
kubectl view gather --must-gather-path=~/my-must-gather logs my-pod
````

## 🔧 Installation

### Manual

1. Download the binary from [Releases](https://github.com/zimnx/kubectl-view-gather/releases)
2. Rename and move it to a directory in your `PATH`:

```sh
mv kubectl-view-gather /usr/local/bin/kubectl-view-gather
chmod +x /usr/local/bin/kubectl-view-gather
```

## 🚀 Usage

First, specify the path to a must-gather dump:

```sh
kubectl view gather --must-gather-path /path/to/must-gather get pods
```

Or set the path via an environment variable:

```sh
export MUST_GATHER_PATH=/path/to/must-gather
kubectl view gather get nodes
```

