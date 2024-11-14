package entry

import (
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/stretchr/testify/assert"
)

func TestEntry_String(t *testing.T) {
	e := Entry{Host: "host", Port: "22", User: "user"}
	assert.Equal(t, "user@host:22", e.String())
}

func TestEntry_Tidy(t *testing.T) {
	patches := gomonkey.NewPatches()
	defer patches.Reset()

	e := &Entry{}
	if err := e.Tidy(); err != nil {
		t.Fatalf("Received unexpected error:\n%+v", err)
	}

	assert.Equal(t, "root", e.User)
	assert.Equal(t, "22", e.Port)
	assert.Equal(t, defaultIdentityFile, e.KeyPath)
}
