# 环境变量

SSX 支持如下环境变量:

|环境变量名| 说明 | 默认值 |
|:---|:---|:---|
|`SSX_DB_PATH`| 用于存储条目的数据库文件 | ~/.ssx.db |
|`SSX_CONNECT_TIMEOUT`| SSH连接超时，单位支持 h/m/s | `10s` |
|`SSX_IMPORT_SSH_CONFIG`| 是否导入用户ssh配置 | |
|`SSX_SECRET_KEY`| [v0.4+ 废弃] 为了兼容旧版本，该参数会等价于 `SSX_DEVICE_ID`  |  |
|`SSX_DEVICE_ID`| 数据库文件需要绑定的设备ID，可以通过设置相同的该环境变量来实现不同设备共用同一份数据库 | [设备ID](#设备id) |

## 解释

### SSX_IMPORT_SSH_CONFIG

这个环境变量不设置时，ssx 默认是不会读取用户的 `~/.ssh/config` 文件的，ssx 只使用自己存储文件进行检索。如果将这个环境变量设置为非空（任意字符串），ssx 就会在初始化的时候加载用户 ssh 配置文件中存在的服务器条目，但 ssx 仅读取用于检索和登录，并不会将这些条目持久化到 ssx 的存储文件中，所以，如果 `ssx IP` 登录时，这个 `IP` 是 `~/.ssh/config` 文件中已经配置过登录验证方式的服务器，ssx 匹配到就直接登录了。但 ssx list 查看时，该服务器会被显示到 `found in ssh config` 的表格中，这个表格中的条目是不具有 ID 属性的。

### 设备ID

- Linux 使用 `/var/lib/dbus/machine-id` ([man](http://man7.org/linux/man-pages/man5/machine-id.5.html))
- OS X 使用 `IOPlatformUUID`
- Windows 使用 `MachineGuid`，取自 `HKEY_LOCAL_MACHINE\SOFTWARE\Microsoft\Cryptography`
