Kube pool manager
=================

[![license](https://img.shields.io/github/license/webdevops/kube-pool-manager.svg)](https://github.com/webdevops/kube-pool-manager/blob/master/LICENSE)
[![DockerHub](https://img.shields.io/badge/DockerHub-webdevops%2Fkube--pool--manager-blue)](https://hub.docker.com/r/webdevops/kube-pool-manager/)
[![Quay.io](https://img.shields.io/badge/Quay.io-webdevops%2Fkube--pool--manager-blue)](https://quay.io/repository/webdevops/kube-pool-manager)

Manager for Kubernetes pool, automatic applies configuration (annotations, labels, configSource, role) to kubernetes nodes based on any node spec.

Supports JSON path, value and regexp matching.

Sets following settings on nodes if matched:
- node role
- node labels
- node annotations
- node [configSource](https://kubernetes.io/docs/tasks/administer-cluster/reconfigure-kubelet/) 

Node settings are applied on startup and for new nodes (delayed until they are ready) and (optional) on watch timeout.

Configuration
-------------

```
Usage:
  kube-pool-manager [OPTIONS]

Application Options:
      --debug                    debug mode [$DEBUG]
  -v, --verbose                  verbose mode [$VERBOSE]
      --log.json                 Switch log output to json format [$LOG_JSON]
      --instance.nodename=       Name of node where autopilot is running [$INSTANCE_NODENAME]
      --instance.namespace=      Name of namespace where autopilot is running [$INSTANCE_NAMESPACE]
      --instance.pod=            Name of pod where autopilot is running [$INSTANCE_POD]
      --kube.node.labelselector= Node Label selector which nodes should be checked [$KUBE_NODE_LABELSELECTOR]
      --kube.watch.timeout=      Timeout & full resync for node watch (time.Duration) (default: 24h) [$KUBE_WATCH_TIMEOUT]
      --kube.watch.reapply       Reapply node settings on watch timeout [$KUBE_WATCH_REAPPLY]
      --lease.enable             Enable lease (leader election; enabled by default in docker images) [$LEASE_ENABLE]
      --lease.name=              Name of lease lock (default: kube-pool-manager-leader) [$LEASE_NAME]
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
