package entry

import (
	"fmt"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/stretchr/testify/assert"
	"github.com/vimiix/ssx/internal/utils"
)

func TestEntry_String(t *testing.T) {
	e := Entry{Host: "host", Port: "22", User: "user"}
	assert.Equal(t, "user@host:22", e.String())
}

func TestEntry_Tidy(t *testing.T) {
	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyFuncReturn(utils.CurrentUserName, "mockuser", nil)
	patches.ApplyFunc(utils.ExpandHomeDir, func(path string) string { return path })

	e := &Entry{}
	if err := e.Tidy(); err != nil {
		t.Fatalf(fmt.Sprintf("Received unexpected error:\n%+v", err))
	}

	assert.Equal(t, "mockuser", e.User)
	assert.Equal(t, "22", e.Port)
	assert.Equal(t, defaultIdentityFile, e.KeyPath)
}
