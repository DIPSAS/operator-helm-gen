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

		p := NewPatcher(Spec{KeepResources: map[string]string{"CustomResourceDefinition": "ok"}})
		err = p.Generate(resources)

		assert.NoError(t, err)
	})
}

func TestGetAnnotationsFromJSONData(t *testing.T) {
	t.Run("get annotations from json data", func(t *testing.T) {
		jsonData := []byte(`{ "apiVersion": "apiextensions.k8s.io/v1", "kind": "CustomResourceDefinition", "metadata": { "annotations": { "controller-gen.kubebuilder.io/version": "v0.19.0", "meta.helm.sh/release-name": "external-secrets"} } }`)

		y, err := yaml.YAMLToJSON(jsonData)
		if assert.NoError(t, err) {
			assert.NotEmpty(t, y)
		}

		expectedAnnotations := make(map[string]string)
		expectedAnnotations["controller-gen.kubebuilder.io/version"] = "v0.19.0"
		expectedAnnotations["meta.helm.sh/release-name"] = "external-secrets"

		annotations, err := getAnnotationsFromJSONData(y)
		if assert.NoError(t, err) {
			assert.Equal(t, len(expectedAnnotations), len(annotations))
			for k, v := range expectedAnnotations {
				annotationFromData, ok := annotations[k]
				if assert.True(t, ok) {
					assert.Equal(t, v, annotationFromData)
				}
			}
		}
	})
}

func TestMergeAnnotations(t *testing.T) {
	testCases := []struct {
		name                string
		inputFilepath       string
		annotationsToAdd    map[string]string
		expectedAnnotations map[string]string
	}{
		{name: "merge annotations on resource with existing annotations", inputFilepath: filepath.Join("..", "..", "test", "operator-helm-gen", "testdata", "crd-with-annotations.yaml"), annotationsToAdd: map[string]string{"argocd.argoproj.io/sync-wave": "123"}},
		{name: "merge annotations to resource that has no annotations", inputFilepath: filepath.Join("..", "..", "test", "operator-helm-gen", "testdata", "crd-without-annotations.yaml"), annotationsToAdd: map[string]string{"argocd.argoproj.io/sync-wave": "123"}},
		{name: "merge annotations to resource that has same annotations does not merge", inputFilepath: filepath.Join("..", "..", "test", "operator-helm-gen", "testdata", "crd-with-annotations.yaml"), annotationsToAdd: map[string]string{"dips.com/testing": "value-from-test"}, expectedAnnotations: map[string]string{"dips.com/testing": "test"}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			data, err := os.ReadFile(tc.inputFilepath)
			if assert.NoError(t, err) {
				assert.NotEmpty(t, data)
			}

			result, err := mergeAnnotations(data, tc.annotationsToAdd)
			if assert.NoError(t, err) {
				assert.NotEmpty(t, result)
			}

			if len(tc.expectedAnnotations) == 0 {
				tc.expectedAnnotations = tc.annotationsToAdd
			}
			annotationsFromResult, err := getAnnotationsFromJSONData(result)
			if assert.NoError(t, err) {
				assert.NotEmpty(t, annotationsFromResult)
				for k, v := range tc.expectedAnnotations {
					val, ok := annotationsFromResult[k]
					if assert.True(t, ok) {
						assert.Equal(t, v, val)
					}
				}
			}
		})
	}
}

func TestAnnotations(t *testing.T) {
	testCases := []struct {
		name                string
		inputFilepath       string
		outputDir           string
		resources           map[string]string
		annotations         map[string]string
		expectedAnnotations map[string]string
	}{
		{name: "add annotations to kustomize output", inputFilepath: filepath.Join("..", "..", "test", "operator-helm-gen", "testdata", "kustomize-output.yaml"), outputDir: "generated_tmp", resources: map[string]string{"CustomResourceDefinition": "yes"}, annotations: map[string]string{"argocd.argoproj.io/sync-wave": "1234"}},
		{name: "add annotations to crd with annotations", inputFilepath: filepath.Join("..", "..", "test", "operator-helm-gen", "testdata", "crd-with-annotations.yaml"), outputDir: "generated_tmp2", resources: map[string]string{"CustomResourceDefinition": "yes"}, annotations: map[string]string{"argocd.argoproj.io/sync-wave": "1234"}},
		{name: "add annotations to crd that has no annotations", inputFilepath: filepath.Join("..", "..", "test", "operator-helm-gen", "testdata", "crd-without-annotations.yaml"), outputDir: "generated_tmp3", resources: map[string]string{"CustomResourceDefinition": "yes"}, annotations: map[string]string{"argocd.argoproj.io/sync-wave": "1234"}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			data, err := os.ReadFile(tc.inputFilepath)
			if assert.NoError(t, err) {
				assert.NotEmpty(t, data)
			}

			outputDir := tc.outputDir
			err = os.Mkdir(outputDir, 0777)
			if assert.NoError(t, err) {
				files, err := os.ReadDir(outputDir)
				if assert.Nil(t, err) {
					assert.Equal(t, 0, len(files))
				}
			}

			resources := GetResources(data)
			assert.NotEmpty(t, resources)

			p := NewPatcher(Spec{KeepResources: tc.resources, OutputDir: outputDir, Annotations: tc.annotations})
			err = p.Generate(resources)

			assert.NoError(t, err)

			if len(tc.expectedAnnotations) == 0 {
				tc.expectedAnnotations = tc.annotations
			}

			generatedFiles, err := os.ReadDir(outputDir)
			if assert.Nil(t, err) {
				assert.NotEqual(t, 0, len(generatedFiles))

				for _, f := range generatedFiles {
					assert.False(t, f.IsDir())

					contents, err := os.ReadFile(filepath.Join(outputDir, f.Name()))
					if assert.NoError(t, err) {
						if assert.NotEmpty(t, contents) {
							j, err := yaml.YAMLToJSON(contents)
							if assert.NoError(t, err) {
								foundAnnotations, err := getAnnotationsFromJSONData(j)
								if assert.NoError(t, err) {
									assert.NotEmpty(t, foundAnnotations)
									for k, v := range tc.expectedAnnotations {
										foundVal, ok := foundAnnotations[k]
										if assert.True(t, ok) {
											assert.Equal(t, v, foundVal)
										}
									}
								}
							}
						}
					}
				}

				_ = os.RemoveAll(outputDir) //ignore error
			}
		})
	}
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

		p := NewPatcher(Spec{KeepResources: keepers, OutputDir: outputDir})
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
