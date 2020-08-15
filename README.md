Kube pool manager
=================

[![license](https://img.shields.io/github/license/webdevops/kube-pool-manager.svg)](https://github.com/webdevops/kube-pool-manager/blob/master/LICENSE)
[![Docker](https://img.shields.io/docker/cloud/automated/webdevops/kube-pool-manager)](https://hub.docker.com/r/webdevops/kube-pool-manager/)
[![Docker Build Status](https://img.shields.io/docker/cloud/build/webdevops/kube-pool-manager)](https://hub.docker.com/r/webdevops/kube-pool-manager/)

Manager for Kubernetes pool, automatic applies configuration (annotations, labels, configSource, role) to kubernetes nodes based on any node spec.

Supports JSON path, value and regexp matching.

Sets following settings on nodes if matched:
- node role
- node labels
- node annotations
- node [configSource](https://kubernetes.io/docs/tasks/administer-cluster/reconfigure-kubelet/) 

Configuration
-------------

```
Usage:
  kube-pool-manager [OPTIONS]

Application Options:
      --debug                    debug mode [$DEBUG]
  -v, --verbose                  verbose mode [$VERBOSE]
      --log.json                 Switch log output to json format [$LOG_JSON]
      --kube.node.labelselector= Node Label selector which nodes should be checked [$KUBE_NODE_LABELSELECTOR]
      --kube.watch.timeout=      Timeout & full resync for node watch (time.Duration) (default: 24h)
                                 [$KUBE_WATCH_TIMEOUT]
      --dry-run                  Dry run (do not apply to nodes) [$DRY_RUN]
      --config=                  Config path [$CONFIG]
      --bind=                    Server address (default: :8080) [$SERVER_BIND]

Help Options:
  -h, --help                     Show this help message
```

see [example.yaml](/example.yaml) for configuration file

Metrics
-------

 (see `:8080/metrics`)

| Metric                         | Description                                     |
|:-------------------------------|:------------------------------------------------|
| `poolmanager_node_pool_status` | Status which pool to which node was applied     |
| `poolmanager_node_applied`     | Timestamp when node confg was set               |

Kubernetes deployment
---------------------

see [deployment](/deployment)
