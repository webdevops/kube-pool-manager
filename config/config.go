package config

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/webdevops/kube-pool-manager/k8s"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/util/jsonpath"
	"regexp"
	"strings"
)

type (
	Config struct {
		Pools []PoolConfig `yaml:"pools"`
	}

	PoolConfig struct {
		Name     string               `yaml:"pool"`
		Continue bool                 `yaml:"continue"`
		Selector []PoolConfigSelector `yaml:"selector"`
		Node     PoolConfigNode       `yaml:"node"`
	}

	PoolConfigSelector struct {
		Path   string  `yaml:"path"`
		jsonPath *jsonpath.JSONPath
		Match  *string `yaml:"match"`
		Regexp *string `yaml:"regexp"`
		regexp *regexp.Regexp
	}

	PoolConfigNode struct {
		Role         *string                     `yaml:"role"`
		ConfigSource *PoolConfigNodeConfigSource `yaml:"configSource"`
		Labels       *map[string]string          `yaml:"labels"`
		Annotations  *map[string]string          `yaml:"annotations"`
	}

	PoolConfigNodeConfigSource struct {
		ConfigMap struct {
			Name             string `yaml:"name" json:"name"`
			Namespace        string `yaml:"namespace" json:"namespace"`
			KubeletConfigKey string `yaml:"kubeletConfigKey" json:"kubeletConfigKey"`
		} `yaml:"configMap" json:"configMap"`
	}
)

func (p *PoolConfig) IsMatchingNode(node *corev1.Node) (bool, error) {
	for num, selector := range p.Selector {
		// auto compile regexp
		if selector.Regexp != nil {
			selector.regexp = regexp.MustCompile(*selector.Regexp)
			p.Selector[num].regexp = selector.regexp
		}

		// auto compile json path
		if selector.jsonPath == nil {
			selector.jsonPath = jsonpath.New(p.Name)
			selector.jsonPath.AllowMissingKeys(true)
			if err := selector.jsonPath.Parse(selector.Path); err != nil {
				return false, err
			}
			p.Selector[num].jsonPath = selector.jsonPath
		}

		values, err := selector.jsonPath.FindResults(node)
		if err != nil {
			return false, err
		}

		if len(values) == 1 && len(values[0]) == 1 {
			val := values[0][0].String()
			selectorMatches := false

			// compare value
			if selector.Match != nil {
				if strings.Compare(val, *selector.Match) == 0 {
					selectorMatches = true
				} else {
					log.Tracef("Node \"%s\": path \"%s\" with value \"%s\" is not matching value \"%s\"", node.Name, selector.Path, val, *selector.Match)
				}
			}

			// regexp
			if selector.regexp != nil {
				if selector.regexp.MatchString(val) {
					selectorMatches = true
				} else {
					log.Tracef("Node \"%s\": path \"%s\" with value \"%s\" is not matching regexp \"%s\"", node.Name, selector.Path, val, *selector.Regexp)
				}
			}

			if !selectorMatches {
				return false, nil
			}
		} else {
			// not found -> not matching
			log.Tracef("Node \"%s\": path \"%s\" not found", node.Name, selector.Path)
			return false, nil
		}
	}

	return true, nil
}

func (p *PoolConfig) CreateJsonPatchSet() (patches []k8s.JsonPatch) {
	patches = []k8s.JsonPatch{}

	if p.Node.Role != nil {
		name := "kubernetes.io/role"
		patches = append(patches, k8s.JsonPatchString{
			Op:    "replace",
			Path:  fmt.Sprintf("/metadata/labels/%s", k8s.PatchPathEsacpe(name)),
			Value: *p.Node.Role,
		})
	}

	if p.Node.ConfigSource != nil {
		patches = append(patches, k8s.JsonPatchObject{
			Op:    "replace",
			Path:  "/spec/configSource",
			Value: *p.Node.ConfigSource,
		})
	}

	if p.Node.Labels != nil {
		for name, value := range *p.Node.Labels {
			patches = append(patches, k8s.JsonPatchString{
				Op:    "replace",
				Path:  fmt.Sprintf("/metadata/labels/%s", k8s.PatchPathEsacpe(name)),
				Value: value,
			})
		}
	}

	if p.Node.Annotations != nil {
		for name, value := range *p.Node.Annotations {
			patches = append(patches, k8s.JsonPatchString{
				Op:    "replace",
				Path:  fmt.Sprintf("/metadata/annotations/%s", k8s.PatchPathEsacpe(name)),
				Value: value,
			})
		}
	}

	return patches
}
