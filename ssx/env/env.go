package env

import (
	"os"
	"strings"
)

const (
	SSXDBPath          = "SSX_DB_PATH"
	SSXConnectTimeout  = "SSX_CONNECT_TIMEOUT"
	SSXImportSSHConfig = "SSX_IMPORT_SSH_CONFIG" // 设置了该环境变量的话，就会自动将 ~/.ssh/config 中的条目也加载
	SSXUnsafeMode      = "SSX_UNSAFE_MODE"       // deprecated
	SSXSecretKey       = "SSX_SECRET_KEY"        // deprecated, replaced by SSX_DEVICE_ID
	SSXDeviceID        = "SSX_DEVICE_ID"
)

func IsUnsafeMode() bool {
	switch strings.ToLower(os.Getenv(SSXUnsafeMode)) {
	case "t", "true", "on", "1":
		return true
	default:
		return false
	}
}
