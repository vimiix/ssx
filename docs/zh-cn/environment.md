# 环境变量

SSX 支持如下环境变量:

|环境变量名| 说明 | 默认值 |
|:---|:---|:---|
|`SSX_DB_PATH`| 用于存储条目的数据库文件 | ~/.ssx.db |
|`SSX_CONNECT_TIMEOUT`| SSH连接超时，单位支持 h/m/s | `10s` |
|`SSX_IMPORT_SSH_CONFIG`| 是否导入用户ssh配置 | |
|`SSX_UNSAFE_MODE`| 密码是否以不安全模式存储，默认会以服务器的设备ID进行加密，如果设置了 `SSX_SECRET_KEY` 则以该值加密| |
|`SSX_SECRET_KEY`| 用于加密条目密码的密钥 | [设备ID](#设备id) |

## 解释

### SSX_IMPORT_SSH_CONFIG

这个环境变量不设置时，ssx 默认是不会读取用户的 `~/.ssh/config` 文件的，ssx 只使用自己存储文件进行检索。如果将这个环境变量设置为非空（任意字符串），ssx 就会在初始化的时候加载用户 ssh 配置文件中存在的服务器条目，但 ssx 仅读取用于检索和登录，并不会将这些条目持久化到 ssx 的存储文件中，所以，如果 `ssx IP` 登录时，这个 `IP` 是 `~/.ssh/config` 文件中已经配置过登录验证方式的服务器，ssx 匹配到就直接登录了。但 ssx list 查看时，该服务器会被显示到 `found in ssh config` 的表格中，这个表格中的条目是不具有 ID 属性的。

### SSX_UNSAFE_MODE

默认情况使用设备ID对条目的密码进行加密，所以此时 `SSX_DB_PATH` 所存储的信息无法直接拷贝迁移到其他机器使用，如果设置了 `SSX_UNSAFE_MODE` 则可以直接迁移到其他机器即可使用。

另外如果为了便于迁移，可以通过在多台机器上设置相同的 `SSX_SECRET_KEY` 环境变量，接口共享条目数据库文件。

### 设备ID

- Linux 使用 `/var/lib/dbus/machine-id` ([man](http://man7.org/linux/man-pages/man5/machine-id.5.html))
- OS X 使用 `IOPlatformUUID`
- Windows 使用 `MachineGuid`，取自 `HKEY_LOCAL_MACHINE\SOFTWARE\Microsoft\Cryptography`
