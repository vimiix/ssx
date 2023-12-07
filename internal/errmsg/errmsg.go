package errmsg

import (
	"github.com/pkg/errors"
)

var (
	ErrEntryNotExist = errors.New("entry does not exist")
	ErrRepoNotOpen   = errors.New("repo is not open")
)
