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
