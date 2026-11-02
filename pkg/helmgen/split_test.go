package helmgen

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetResources(t *testing.T) {
	t.Run("test empty data", func(t *testing.T) {
		var data []byte

		res := GetResources(data)
		assert.NotNil(t, res)
		assert.Empty(t, res)
	})

	t.Run("test empty input", func(t *testing.T) {
		i := ""
		data := []byte(i)

		res := GetResources(data)
		assert.NotNil(t, res)
		assert.Empty(t, res)
	})

	t.Run("test data with single element", func(t *testing.T) {
		input := "abc"
		data := []byte(input)
		res := GetResources(data)

		if assert.NotEmpty(t, res) {
			assert.Equal(t, 1, len(res))
		}
	})

	t.Run("test byte slice with more than one element", func(t *testing.T) {
		input := "abc\n---\n123"
		data := []byte(input)
		res := GetResources(data)

		if assert.NotEmpty(t, res) {
			assert.Equal(t, 2, len(res))
		}
	})
}

func TestGetAnnotationsFromArg(t *testing.T) {
	t.Run("parse annotations and return map", func(t *testing.T) {
		arg := "argocd.argoproj.io/sync-wave=1234,argocd.argoproj.io/sync-options=Delete=false"
		annotations, err := GetAnnotationsFromArg(arg)
		if assert.NoError(t, err) {
			assert.NotEmpty(t, annotations)
			assert.Equal(t, 2, len(annotations))
			assert.Equal(t, "1234", annotations["argocd.argoproj.io/sync-wave"])
			assert.Equal(t, "Delete=false", annotations["argocd.argoproj.io/sync-options"])
		}
	})

	t.Run("invalid annotation raises error", func(t *testing.T) {
		arg := "argocd.argoproj.io/sync-wave=1234,argocd.argoproj.io/sync-options"
		annotations, err := GetAnnotationsFromArg(arg)
		if assert.Error(t, err) {
			assert.Nil(t, annotations)
		}
	})
}
