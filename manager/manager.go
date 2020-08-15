package manager

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/operator-framework/operator-sdk/pkg/leader"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"github.com/webdevops/kube-pool-manager/config"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"time"
)

type (
	KubePoolManager struct {
		Opts   config.Opts
		Config config.Config

		ctx       context.Context
		k8sClient *kubernetes.Clientset

		prometheus struct {
			nodePoolStatus *prometheus.GaugeVec
			nodeApplied    *prometheus.GaugeVec
		}
	}
)

func (m *KubePoolManager) Init() {
	m.ctx = context.Background()
	m.initK8s()
	m.initPrometheus()
}

func (r *KubePoolManager) initPrometheus() {
	r.prometheus.nodePoolStatus = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "poolmanager_node_pool_status",
			Help: "kube-pool-manager node pool config status",
		},
		[]string{"nodeName", "pool"},
	)
	prometheus.MustRegister(r.prometheus.nodePoolStatus)

	r.prometheus.nodeApplied = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "poolmanager_node_applied",
			Help: "kube-pool-manager node applied time",
		},
		[]string{"nodeName"},
	)
	prometheus.MustRegister(r.prometheus.nodeApplied)
}

func (r *KubePoolManager) initK8s() {
	var err error
	var config *rest.Config

	if kubeconfig := os.Getenv("KUBECONFIG"); kubeconfig != "" {
		// KUBECONFIG
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			panic(err.Error())
		}
	} else {
		// K8S in cluster
		config, err = rest.InClusterConfig()
		if err != nil {
			panic(err.Error())
		}
	}

	r.k8sClient, err = kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
}

func (m *KubePoolManager) Start() {
	go func() {
		m.leaderElect()
		for {
			log.Info("(re)starting node watch")
			if err := m.startNodeWatch(); err != nil {
				log.Errorf("node watcher stopped: %v", err)
			}
		}
	}()
}

func (m *KubePoolManager) leaderElect() {
	if m.Opts.Lease.Enabled {
		log.Info("trying to become leader")
		if m.Opts.Instance.Pod != nil && os.Getenv("POD_NAME") == "" {
			err := os.Setenv("POD_NAME", *m.Opts.Instance.Pod)
			if err != nil {
				log.Panic(err)
			}
		}

		time.Sleep(15 * time.Second)
		err := leader.Become(m.ctx, m.Opts.Lease.Name)
		if err != nil {
			log.Error(err, "Failed to retry for leader lock")
			os.Exit(1)
		}
		log.Info("aquired leader lock, continue")
	}
}

func (m *KubePoolManager) startNodeWatch() error {
	timeout := int64(m.Opts.K8s.WatchTimeout.Seconds())
	watchOpts := metav1.ListOptions{
		LabelSelector:  m.Opts.K8s.NodeLabelSelector,
		TimeoutSeconds: &timeout,
		Watch:          true,
	}
	nodeWatcher, err := m.k8sClient.CoreV1().Nodes().Watch(m.ctx, watchOpts)
	if err != nil {
		log.Panic(err)
	}
	defer nodeWatcher.Stop()

	for res := range nodeWatcher.ResultChan() {
		switch res.Type {
		case watch.Added:
			if node, ok := res.Object.(*corev1.Node); ok {
				m.applyNode(node)
			}
		case watch.Error:
			log.Errorf("go watch error event %v", res.Object)
		}
	}

	return fmt.Errorf("terminated")
}

func (m *KubePoolManager) applyNode(node *corev1.Node) {
	contextLogger := log.WithField("node", node.Name)

	for _, poolConfig := range m.Config.Pools {
		m.prometheus.nodePoolStatus.WithLabelValues(node.Name, poolConfig.Name).Set(0)
		poolLogger := contextLogger.WithField("pool", poolConfig.Name)
		matching, err := poolConfig.IsMatchingNode(node)
		if err != nil {
			poolLogger.Panic(err)
		}

		if matching {
			poolLogger.Infof("applying pool \"%s\" to node \"%s\"", poolConfig.Name, node.Name)

			// create json patch
			patchSet := poolConfig.CreateJsonPatchSet()
			patchBytes, patchErr := json.Marshal(patchSet)
			if patchErr != nil {
				poolLogger.Errorf("failed to create json patch: %v", err)
				return
			}

			if !m.Opts.DryRun {
				// patch node
				_, k8sError := m.k8sClient.CoreV1().Nodes().Patch(m.ctx, node.Name, types.JSONPatchType, patchBytes, metav1.PatchOptions{})
				if k8sError != nil {
					poolLogger.Errorf("failed to apply json patch: %v", k8sError)
					return
				}
			} else {
				poolLogger.Infof("Not applying pool config, dry-run active")
			}

			m.prometheus.nodePoolStatus.WithLabelValues(node.Name, poolConfig.Name).Set(1)
			m.prometheus.nodeApplied.WithLabelValues(node.Name).SetToCurrentTime()

			// check if this more pool configurations should be applied
			if !poolConfig.Continue {
				break
			}
		} else {
			poolLogger.Debugf("Node NOT matches pool configuration \"%s\"", poolConfig.Name)
		}
	}
}
