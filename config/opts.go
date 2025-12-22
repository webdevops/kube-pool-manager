package config

import (
	"encoding/json"
	"time"
)

type (
	Opts struct {
		// logger
		Logger struct {
			Level  string `long:"log.level"    env:"LOG_LEVEL"   description:"Log level" choice:"trace" choice:"debug" choice:"info" choice:"warning" choice:"error" default:"info"`                          // nolint:staticcheck // multiple choices are ok
			Format string `long:"log.format"   env:"LOG_FORMAT"  description:"Log format" choice:"logfmt" choice:"json" default:"logfmt"`                                                                     // nolint:staticcheck // multiple choices are ok
			Source string `long:"log.source"   env:"LOG_SOURCE"  description:"Show source for every log message (useful for debugging and bug reports)" choice:"" choice:"short" choice:"file" choice:"full"` // nolint:staticcheck // multiple choices are ok
			Color  string `long:"log.color"    env:"LOG_COLOR"   description:"Enable color for logs" choice:"" choice:"auto" choice:"yes" choice:"no"`                                                        // nolint:staticcheck // multiple choices are ok
			Time   bool   `long:"log.time"     env:"LOG_TIME"    description:"Show log time"`
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
