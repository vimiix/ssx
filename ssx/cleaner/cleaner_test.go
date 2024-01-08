package cleaner

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClean(t *testing.T) {
	var called bool
	cb := func() {
		called = true
	}
	assert.Equal(t, 0, len(cbs))
	RegisterCallback(cb)
	assert.Equal(t, 1, len(cbs))

	Clean()
	assert.True(t, called)
}
