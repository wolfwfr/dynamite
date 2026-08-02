package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSingleCharMatch(t *testing.T) {
	t.Run("matches against single character", func(t *testing.T) {
		res := SingleChar.Match([]byte("q"))
		assert.True(t, res)
	})
	t.Run("does not match against multiple character", func(t *testing.T) {
		res := SingleChar.Match([]byte("esc"))
		assert.False(t, res)
	})
}
