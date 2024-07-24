package helmgen

import (
	"os"
	"path/filepath"
	"testing"

	jsonIterator "github.com/json-iterator/go"
	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/yaml"
)

func TestPatchNamespace(t *testing.T) {
	t.Run("static namespace element gets templated", func(t *testing.T) {
		data, err := os.ReadFile(filepath.Join("..", "..", "test", "operator-helm-gen", "testdata", "rolebinding.yaml"))

		if assert.NoError(t, err) {
			assert.NotEmpty(t, data)
		}

		j, err := yaml.YAMLToJSON(data)
		if assert.NoError(t, err) {
			assert.NotEmpty(t, j)
		}

		resourceNS := jsonIterator.Get(j, "metadata", "namespace").ToString()
		assert.NotEmpty(t, resourceNS)
		assert.NotContains(t, resourceNS, ".Release.Namespace")

		patched, err := patchResource(j)
		if assert.NoError(t, err) {
			assert.NotEmpty(t, patched)
		}

		patchedResourceNS := jsonIterator.Get(patched, "metadata", "namespace").ToString()
		assert.NotEmpty(t, patchedResourceNS)
		assert.Contains(t, patchedResourceNS, ".Release.Namespace")
	})
}

func TestParseCRD(t *testing.T) {
	t.Run("crd with escape chars in descriptions", func(t *testing.T) {
		data, err := os.ReadFile(filepath.Join("..", "..", "test", "operator-helm-gen", "testdata", "kustomize-output.yaml"))

		if assert.NoError(t, err) {
			assert.NotEmpty(t, data)
		}

		resources := GetResources(data)
		assert.NotEmpty(t, resources)

		p := NewPatcher(Spec{KeepRresources: map[string]string{"CustomResourceDefinition": "ok"}})
		err = p.Generate(resources)

		assert.NoError(t, err)
	})
}
