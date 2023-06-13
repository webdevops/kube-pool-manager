package config

import (
	"encoding/json"
	"time"
)

type (
	Opts struct {
		// logger
		Logger struct {
			Debug       bool `long:"log.debug"    env:"LOG_DEBUG"  description:"debug mode"`
			Development bool `long:"log.devel"    env:"LOG_DEVEL"  description:"development mode"`
			Json        bool `long:"log.json"     env:"LOG_JSON"   description:"Switch log output to json format"`
		}

		// instance
		Instance struct {
			Nodename  *string `long:"instance.nodename"    env:"INSTANCE_NODENAME"   description:"Name of node where autopilot is running"`
			Namespace *string `long:"instance.namespace"   env:"INSTANCE_NAMESPACE"   description:"Name of namespace where autopilot is running"`
			Pod       *string `long:"instance.pod"         env:"INSTANCE_POD"         description:"Name of pod where autopilot is running"`
		}

		K8s struct {
			NodeLabelSelector     string        `long:"kube.node.labelselector"     env:"KUBE_NODE_LABELSELECTOR"     description:"Node Label selector which nodes should be checked"        default:""`
			WatchTimeout          time.Duration `long:"kube.watch.timeout"          env:"KUBE_WATCH_TIMEOUT"          description:"Timeout & full resync for node watch (time.Duration)"     default:"24h"`
			ReapplyOnWatchTimeout bool          `long:"kube.watch.reapply"          env:"KUBE_WATCH_REAPPLY"          description:"Reapply node settings on watch timeout"`
		}

		// lease
		Lease struct {
			Enabled bool   `long:"lease.enable"  env:"LEASE_ENABLE"  description:"Enable lease (leader election; enabled by default in docker images)"`
			Name    string `long:"lease.name"    env:"LEASE_NAME"    description:"Name of lease lock"     default:"kube-pool-manager-leader"`
		}

		// server
		Server struct {
			// general options
			Bind         string        `long:"server.bind"              env:"SERVER_BIND"           description:"Server address"        default:":8080"`
			ReadTimeout  time.Duration `long:"server.timeout.read"      env:"SERVER_TIMEOUT_READ"   description:"Server read timeout"   default:"5s"`
			WriteTimeout time.Duration `long:"server.timeout.write"     env:"SERVER_TIMEOUT_WRITE"  description:"Server write timeout"  default:"10s"`
		}

		// general options
		DryRun bool   `long:"dry-run"  env:"DRY_RUN"       description:"Dry run (do not apply to nodes)"`
		Config string `long:"config"   env:"CONFIG"        description:"Config path"        required:"true"`
	}
)

func (o *Opts) GetJson() []byte {
	jsonBytes, err := json.Marshal(o)
	if err != nil {
		panic(err)
	}
	return jsonBytes
}
