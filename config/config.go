package config

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/webdevops/go-common/log/slogger"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/util/jsonpath"

	"github.com/webdevops/kube-pool-manager/k8s"
)

type (
	Config struct {
		Pools []PoolConfig `json:"pools"`
	}

	PoolConfig struct {
		Name     string               `json:"pool"`
		Continue bool                 `json:"continue"`
		Selector []PoolConfigSelector `json:"selector"`
		Node     PoolConfigNode       `json:"node"`
	}

	PoolConfigSelector struct {
		Path     string `json:"path"`
		jsonPath *jsonpath.JSONPath
		Match    *string `json:"match"`
		Regexp   *string `json:"regexp"`
		regexp   *regexp.Regexp
	}

	PoolConfigNode struct {
		Roles        PoolConfigNodeValueMap      `json:"roles"`
		JsonPatches  []k8s.JsonPatchObject       `json:"jsonPatches"`
		ConfigSource *PoolConfigNodeConfigSource `json:"configSource"`
		Labels       PoolConfigNodeValueMap      `json:"labels"`
		Annotations  PoolConfigNodeValueMap      `json:"annotations"`
	}

	PoolConfigNodeConfigSource struct {
		ConfigMap struct {
			Name             string `json:"name"`
			Namespace        string `json:"namespace"`
			KubeletConfigKey string `json:"kubeletConfigKey"`
		} `json:"configMap"`
	}

	PoolConfigNodeValueMap struct {
		entries *map[string]*string
	}
)

func (valueMap *PoolConfigNodeValueMap) Entries() map[string]*string {
	var mapList map[string]*string

	if valueMap.entries != nil {
		mapList = *valueMap.entries
	}

	return mapList
}

func (valueMap *PoolConfigNodeValueMap) UnmarshalJSON(b []byte) error {
	mapList := map[string]*string{}
	err := json.Unmarshal(b, &mapList)
	if err != nil {
		var stringList []string
		err := json.Unmarshal(b, &stringList)
		if err != nil {
			return err
		}
		if len(stringList) > 0 {
			emptyVal := ""
			for _, val := range stringList {
				mapList[val] = &emptyVal
			}
			valueMap.entries = &mapList
		}
	} else {
		valueMap.entries = &mapList
	}
	return nil
}

func (p *PoolConfig) IsMatchingNode(logger *slogger.Logger, node *corev1.Node) (bool, error) {
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
					logger.Debugf("Node \"%s\": path \"%s\" with value \"%s\" is not matching value \"%s\"", node.Name, selector.Path, val, *selector.Match)
				}
			}

			// regexp
			if selector.regexp != nil {
				if selector.regexp.MatchString(val) {
					selectorMatches = true
				} else {
					logger.Debugf("Node \"%s\": path \"%s\" with value \"%s\" is not matching regexp \"%s\"", node.Name, selector.Path, val, *selector.Regexp)
				}
			}

			if !selectorMatches {
				return false, nil
			}
		} else {
			// not found -> not matching
			logger.Debugf("Node \"%s\": path \"%s\" not found", node.Name, selector.Path)
			return false, nil
		}
	}

	return true, nil
}

func (p *PoolConfig) CreateJsonPatchSet(node *corev1.Node) (patchSet *k8s.JsonPatchSet) {
	patchSet = k8s.NewJsonPatchSet()

	// node roles
	for roleName, roleValue := range p.Node.Roles.Entries() {
		label := fmt.Sprintf("node-role.kubernetes.io/%s", roleName)
		if roleValue != nil {
			value := *roleValue
			patchSet.Add(k8s.JsonPatchString{
				Op:    "replace",
				Path:  fmt.Sprintf("/metadata/labels/%s", k8s.PatchPathEsacpe(label)),
				Value: &value,
			})
		} else if _, labelExists := node.Labels[label]; labelExists {
			patchSet.Add(k8s.JsonPatchString{
				Op:   "remove",
				Path: fmt.Sprintf("/metadata/labels/%s", k8s.PatchPathEsacpe(label)),
			})
		}
	}

	// node config source
	if p.Node.ConfigSource != nil {
		patchSet.Add(k8s.JsonPatchObject{
			Op:    "replace",
			Path:  "/spec/configSource",
			Value: *p.Node.ConfigSource,
		})
	}

	// node labels
	for labelName, labelValue := range p.Node.Labels.Entries() {
		if labelValue != nil {
			value := *labelValue
			patchSet.Add(k8s.JsonPatchString{
				Op:    "replace",
				Path:  fmt.Sprintf("/metadata/labels/%s", k8s.PatchPathEsacpe(labelName)),
				Value: &value,
			})
		} else if _, labelExists := node.Labels[labelName]; labelExists {
			patchSet.Add(k8s.JsonPatchString{
				Op:   "remove",
				Path: fmt.Sprintf("/metadata/labels/%s", k8s.PatchPathEsacpe(labelName)),
			})
		}
	}

	// node annotations
	for annotationName, annotationValue := range p.Node.Annotations.Entries() {
		if annotationValue != nil {
			value := *annotationValue
			patchSet.Add(k8s.JsonPatchString{
				Op:    "replace",
				Path:  fmt.Sprintf("/metadata/annotations/%s", k8s.PatchPathEsacpe(annotationName)),
				Value: &value,
			})
		} else if _, labelExists := node.Annotations[annotationName]; labelExists {
			patchSet.Add(k8s.JsonPatchString{
				Op:   "remove",
				Path: fmt.Sprintf("/metadata/annotations/%s", k8s.PatchPathEsacpe(annotationName)),
			})
		}
	}

	// custom patches
	for _, patch := range p.Node.JsonPatches {
		patchSet.Add(patch)
	}

	return
}
