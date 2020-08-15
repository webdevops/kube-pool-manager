package k8s

import (
	"encoding/json"
	"strings"
)

type (
	JsonPatch interface{}

	JsonPatchString struct {
		JsonPatch
		Op    string `json:"op"`
		Path  string `json:"path"`
		Value string `json:"value"`
	}

	JsonPatchObject struct {
		JsonPatch
		Op    string      `json:"op"`
		Path  string      `json:"path"`
		Value interface{} `json:"value"`
	}

	JsonPatchSet struct {
		List map[string]JsonPatch
	}
)

func PatchPathEsacpe(val string) string {
	val = strings.ReplaceAll(val, "~", "~0")
	val = strings.ReplaceAll(val, "/", "~1")
	return val
}

func NewJsonPatchSet() *JsonPatchSet {
	set := JsonPatchSet{}
	set.List = map[string]JsonPatch{}
	return &set
}

func (set *JsonPatchSet) AddSet(patchSet *JsonPatchSet) {
	for _, patch := range patchSet.List {
		set.Add(patch)
	}
}

func (set *JsonPatchSet) Add(patch JsonPatch) {
	switch v := patch.(type) {
	case JsonPatchString:
		set.List[v.Path] = patch
	case JsonPatchObject:
		set.List[v.Path] = patch
	default:
		panic("jsonPatch type not defined or allowed")
	}
}

func (set *JsonPatchSet) Marshal() ([]byte, error) {
	patchList := []JsonPatch{}
	for _, patch := range set.List {
		patchList = append(patchList, patch)
	}

	return json.Marshal(patchList)
}
