package config

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"time"
)

type (
	Opts struct {
		// logger
		Logger struct {
			Debug   bool `           long:"debug"        env:"DEBUG"    description:"debug mode"`
			Verbose bool `short:"v"  long:"verbose"      env:"VERBOSE"  description:"verbose mode"`
			LogJson bool `           long:"log.json"     env:"LOG_JSON" description:"Switch log output to json format"`
		}

		// instance
		Instance struct {
			Nodename  *string `long:"instance.nodename"    env:"INSTANCE_NODENAME"   description:"Name of node where autopilot is running"`
			Namespace *string `long:"instance.namespace"   env:"INSTANCE_NAMESPACE"   description:"Name of namespace where autopilot is running"`
			Pod       *string `long:"instance.pod"         env:"INSTANCE_POD"         description:"Name of pod where autopilot is running"`
		}

		K8s struct {
			NodeLabelSelector string        `long:"kube.node.labelselector"     env:"KUBE_NODE_LABELSELECTOR"     description:"Node Label selector which nodes should be checked"        default:""`
			WatchTimeout      time.Duration `long:"kube.watch.timeout"          env:"KUBE_WATCH_TIMEOUT"          description:"Timeout & full resync for node watch (time.Duration)"     default:"24h"`
		}

		// lease
		Lease struct {
			Enabled bool   `long:"lease.enable"  env:"LEASE_ENABLE"  description:"Enable lease (leader election; enabled by default in docker images)"`
			Name    string `long:"lease.name"    env:"LEASE_NAME"    description:"Name of lease lock"     default:"kube-pool-manager-leader"`
		}

		// general options
		DryRun     bool   `long:"dry-run"  env:"DRY_RUN"       description:"Dry run (do not apply to nodes)"`
		Config     string `long:"config"   env:"CONFIG"        description:"Config path"        required:"true"`
		ServerBind string `long:"bind"     env:"SERVER_BIND"   description:"Server address"     default:":8080"`
	}
)

func (o *Opts) GetJson() []byte {
	jsonBytes, err := json.Marshal(o)
	if err != nil {
		log.Panic(err)
	}
	return jsonBytes
}
