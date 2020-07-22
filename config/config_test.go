package config

import (
	corev1 "k8s.io/api/core/v1"
	"testing"
)

func stringPtr(val string) *string {
	return &val
}

func buildNode() *corev1.Node {
	node := corev1.Node{}
	node.Spec.ProviderID = "azure:///subscriptions/d86bcf13-ddf7-45ea-82f1-6f656767a318/resourceGroups/mc_k8s_mblaschke_westeurope/providers/Microsoft.Compute/virtualMachineScaleSets/aks-agents-35471996-vmss/virtualMachines/30"
	node.ObjectMeta.Annotations = map[string]string{
		"node.kubernetes.io/foobar": "barfoo",
	}
	node.ObjectMeta.Labels = map[string]string{
		"node.kubernetes.io/role": "worker",
	}

	return &node
}

func Test_NodeMatcher(t *testing.T) {
	node := buildNode()

	pool := PoolConfig{
		Name: "testing",
		Selector: []PoolConfigSelector{
			{
				Path:  "{.spec.providerID}",
				Match: stringPtr("azure:///subscriptions/d86bcf13-ddf7-45ea-82f1-6f656767a318/resourceGroups/mc_k8s_mblaschke_westeurope/providers/Microsoft.Compute/virtualMachineScaleSets/aks-agents-35471996-vmss/virtualMachines/30"),
			},
		},
	}
	matching, err := pool.IsMatchingNode(node)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !matching {
		t.Error("Expected not matching, but matching node")
	}

	pool.Selector[0].Match = stringPtr("azure:///subscriptions/d86bcf13-ddf7-45ea-82f1-6f656767a318/resourceGroups/mc_k8s_mblaschke_westeurope/providers/Microsoft.Compute/virtualMachineScaleSets/aks-agents-35471996-vmss/virtualMachines/31")
	matching, err = pool.IsMatchingNode(node)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if matching {
		t.Error("Expected matching, but not matching node")
	}
}

func Test_NodeRegexp(t *testing.T) {
	node := buildNode()

	pool := PoolConfig{
		Name: "testing",
		Selector: []PoolConfigSelector{
			{
				Path:   "{.spec.providerID}",
				Regexp: stringPtr("^.+/resourceGroups/mc_k8s_mblaschke_westeurope/.+$"),
			},
		},
	}
	matching, err := pool.IsMatchingNode(node)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !matching {
		t.Error("Expected not matching, but matching node")
	}
}

func Test_NodeLabelMatcher(t *testing.T) {
	node := buildNode()

	pool := PoolConfig{
		Name: "testing",
		Selector: []PoolConfigSelector{
			{
				Path:  "{.metadata.labels.node\\.kubernetes\\.io/role}",
				Match: stringPtr("worker"),
			},
		},
	}
	matching, err := pool.IsMatchingNode(node)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !matching {
		t.Error("Expected matching, but not matching node")
	}
}
