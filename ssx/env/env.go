package env

const (
	SSXDBPath          = "SSX_DB_PATH"
	SSXConnectTimeout  = "SSX_CONNECT_TIMEOUT"
	SSXImportSSHConfig = "SSX_IMPORT_SSH_CONFIG" // 设置了该环境变量的话，就会自动将 ~/.ssh/config 中的条目也加载
)
