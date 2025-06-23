package helmgen

import (
	"os"
	"path/filepath"
	"strings"
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
func TestParseAll(t *testing.T) {
	t.Run("crd with escape chars in descriptions", func(t *testing.T) {
		data, err := os.ReadFile(filepath.Join("..", "..", "test", "operator-helm-gen", "testdata", "kustomize-output.yaml"))

		if assert.NoError(t, err) {
			assert.NotEmpty(t, data)
		}

		outputDir := "generated_tmp"

		err = os.Mkdir(outputDir, 0777)
		if assert.Nil(t, err) {
			files, err := os.ReadDir(outputDir)
			if assert.Nil(t, err) {
				assert.Equal(t, 0, len(files))
			}
		}

		resources := GetResources(data)
		assert.NotEmpty(t, resources)

		keepers := make(map[string]string)
		keepers["CustomResourceDefinition"] = "ok"
		keepers["ClusterRoleBinding"] = "ok"
		keepers["RoleBinding"] = "ok"
		keepers["Role"] = "ok"
		keepers["ClusterRole"] = "ok"

		p := NewPatcher(Spec{KeepRresources: keepers, OutputDir: outputDir})
		err = p.Generate(resources)

		assert.NoError(t, err)

		generatedFiles, err := os.ReadDir(outputDir)
		if assert.Nil(t, err) {
			assert.NotEqual(t, 0, len(generatedFiles))

			for _, f := range generatedFiles {
				assert.False(t, f.IsDir())

				contents, err := os.ReadFile(filepath.Join(outputDir, f.Name()))
				if assert.NoError(t, err) {
					if assert.NotEmpty(t, contents) {
						y := string(contents)
						assert.False(t, strings.HasPrefix(y, "\n"), "did not expect file %s to start with newline", f.Name())
					}
				}
			}

			_ = os.RemoveAll(outputDir) //ignore error
		}
	})
}
