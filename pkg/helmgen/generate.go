package helmgen

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	jsonpatch "github.com/evanphx/json-patch"
	jsonIterator "github.com/json-iterator/go"
	"sigs.k8s.io/yaml"
)

type Patcher struct {
	Spec Spec
}

func NewPatcher(spec Spec) *Patcher {
	return &Patcher{Spec: spec}
}

func (p *Patcher) Generate(resources []string) error {
	if len(resources) == 0 {
		return nil
	}

	var err error
	for i, resource := range resources {
		err = p.doResource(resource)
		if err != nil {
			return fmt.Errorf("failed parsing element at index %d: %v", i, err)
		}
	}

	return nil
}

func (p *Patcher) doResource(resource string) error {
	if len(resource) == 0 {
		return nil
	}

	b := []byte(resource)

	y, err := yaml.YAMLToJSON(b)
	if err != nil {
		return fmt.Errorf("failed converting YAML to JSON: %w", err)
	}

	resourceName := jsonIterator.Get(y, "metadata", "name")
	resourceKind := jsonIterator.Get(y, "kind")

	name := resourceName.ToString()
	if len(name) == 0 {
		return nil
	}
	// fmt.Printf("resourceName: %v\n\n\n", a)

	kind := resourceKind.ToString()
	if len(kind) == 0 {
		return nil
	}

	_, ok := p.Spec.KeepRresources[kind]
	if !ok {
		return nil
	}

	filename := fmt.Sprintf("%s-%s.yaml", strings.ToLower(kind), strings.ToLower(strings.Replace(name, ".", "-", -1)))
	//fmt.Printf("#filename %s\n", filename)

	patched, err := patchResource(y)
	if err != nil {
		return fmt.Errorf("failed patching resource kind %s, name %s: %v", kind, name, err)
	}

	yamlData, err := yaml.JSONToYAML(patched)
	if err != nil {
		return fmt.Errorf("could not convert patched JSON to YAML: %v", err)
	}

	contents := make([]string, 0)
	contents = append(contents, fmt.Sprintf("#filename %s", filename))
	contents = append(contents, string(yamlData))

	result := strings.Join(contents, "\n")

	if len(p.Spec.OutputDir) > 0 {
		outputFile := filepath.Join(p.Spec.OutputDir, filename)
		out := []byte(result)
		err = os.WriteFile(outputFile, out, os.ModePerm)
		if err != nil {
			return fmt.Errorf("could not write result to output file %s: %v", outputFile, err)
		}
	} else {
		_, err = os.Stdout.WriteString(result)
		if err != nil {
			return fmt.Errorf("could not write result to stdout: %v", err)
		}
	}

	return nil
}

func patchResource(data []byte) ([]byte, error) {

	resourceKind := jsonIterator.Get(data, "kind")
	resourceNS := jsonIterator.Get(data, "metadata", "namespace")

	kind := resourceKind.ToString()
	ns := resourceNS.ToString()

	var r = data

	if len(ns) > 0 {
		nsPatchSpec := []byte(`[
			{"op": "replace", "path": "/metadata/namespace", "value": "{{ .Release.Namespace }}"}
		]`)

		patchNS, err := jsonpatch.DecodePatch(nsPatchSpec)
		if err != nil {
			return nil, fmt.Errorf("failed decoding JSON namespace patch: %w", err)
		}

		r, err = patchNS.Apply(r)
		if err != nil {
			return nil, fmt.Errorf("failed applying JSON namespace patch: %w", err)
		}
	}

	var err error

	switch kind {
	case "RoleBinding":
		r, err = handleRoleBinding(r)
		if err != nil {
			return nil, fmt.Errorf("could not handle RoleBinding YAML: %w", err)
		}

	case "ClusterRoleBinding":
		r, err = handleRoleBinding(r)
		if err != nil {
			return nil, fmt.Errorf("could not handle ClusterRoleBinding YAML: %w", err)
		}
	}

	return r, nil
}

func handleRoleBinding(jsonData []byte) ([]byte, error) {
	subjects := jsonIterator.Get(jsonData, "subjects", '*')
	num := subjects.Size()
	//	fmt.Printf("#valuetype size %v\n", num)

	if num == 0 {
		return jsonData, nil
	}

	for i := 0; i < num; i++ {
		subjectNS := jsonIterator.Get(jsonData, "subjects", i, "namespace")
		if len(subjectNS.ToString()) > 0 {
			subjectSpec := fmt.Sprintf(`[
		{"op": "replace", "path": "/subjects/%d/namespace", "value": "{{ .Release.Namespace }}"}
		]`, i)

			subjectPatch := []byte(subjectSpec)
			patchNS, err := jsonpatch.DecodePatch(subjectPatch)
			if err != nil {
				return nil, fmt.Errorf("failed decoding JSON patch for RoleBinding.subjects.%d.namespace: %w", i, err)
			}

			jsonData, err = patchNS.Apply(jsonData)
			if err != nil {
				return nil, fmt.Errorf("failed applying JSON patch for RoleBinding.subjects.%d.namespace: %w", i, err)
			}
		}

		subjectKind := jsonIterator.Get(jsonData, "subjects", i, "kind")
		if subjectKind.ToString() == "ServiceAccount" {
			kindSpec := fmt.Sprintf(`[
		{"op": "replace", "path": "/subjects/%d/name", "value": "{{ .Values.serviceAccount.name}}"}
		]`, i)

			kindPatch := []byte(kindSpec)
			patchKind, err := jsonpatch.DecodePatch(kindPatch)
			if err != nil {
				return nil, fmt.Errorf("failed decoding JSON patch for RoleBinding.subjects.%d.name: %w", i, err)
			}

			jsonData, err = patchKind.Apply(jsonData)
			if err != nil {
				return nil, fmt.Errorf("failed applying JSON patch for RoleBinding.subjects.%d.namespace: %w", i, err)
			}

		}
	}

	return jsonData, nil
}
