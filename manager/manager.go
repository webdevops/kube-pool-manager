package manager

import (
	"context"
	"encoding/json"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"github.com/webdevops/kube-pool-manager/config"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"strings"
)

type (
	KubePoolManager struct {
		Opts   config.Opts
		Config config.Config

		ctx       context.Context
		k8sClient *kubernetes.Clientset

		prometheus struct {
			nodeApplied *prometheus.GaugeVec
		}
	}
)

func (m *KubePoolManager) Init() {
	m.ctx = context.Background()
	m.initK8s()
	m.initPrometheus()
}

func (r *KubePoolManager) initPrometheus() {
	r.prometheus.nodeApplied = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "poolmanager_node_applied",
			Help: "kube-pool-manager node status",
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
	m.startWatch()
}

func (m *KubePoolManager) startWatch() {
	watch, err := m.k8sClient.CoreV1().Nodes().Watch(m.ctx, metav1.ListOptions{Watch: true})
	if err != nil {
		log.Panic(err)
	}

	go func() {
		for res := range watch.ResultChan() {
			switch strings.ToLower(string(res.Type)) {
			case "added":
				if node, ok := res.Object.(*corev1.Node); ok {
					m.applyNode(node)
				}
			}
		}
	}()
}

func (m *KubePoolManager) applyToAll() error {
	result, err := m.k8sClient.CoreV1().Nodes().List(m.ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, node := range result.Items {
		m.applyNode(&node)
	}

	return nil
}

func (m *KubePoolManager) applyNode(node *corev1.Node) {
	contextLogger := log.WithField("node", node.Name)

	m.prometheus.nodeApplied.WithLabelValues(node.Name).Set(0)

	for _, poolConfig := range m.Config.Pools {
		matching, err := poolConfig.IsMatchingNode(node)
		if err != nil {
			log.Panic(err)
		}

		if matching {
			contextLogger.Infof("Node \"%s\" matches pool configuration \"%s\", applying pool config", node.Name, poolConfig.Name)

			// create json patch
			patchSet := poolConfig.CreateJsonPatchSet()
			patchBytes, patchErr := json.Marshal(patchSet)
			if patchErr != nil {
				contextLogger.Errorf("failed to create json patch: %v", err)
				return
			}

			if !m.Opts.DryRun {
				// patch node
				_, k8sError := m.k8sClient.CoreV1().Nodes().Patch(m.ctx, node.Name, types.JSONPatchType, patchBytes, metav1.PatchOptions{})
				if k8sError != nil {
					contextLogger.Errorf("failed to apply json patch: %v", k8sError)
					return
				}
			} else {
				contextLogger.Infof("Not applying pool config, dry-run active")
			}

			m.prometheus.nodeApplied.WithLabelValues(node.Name).Set(1)
		} else {
			contextLogger.Debugf("Node NOT matches pool configuration \"%s\"", poolConfig.Name)
		}
	}
}
