# 使用方法

## 新增条目

正确登录一次即代表新增条目。

```bash
ssx [-J PROXY_USER@PROXY_HOST:PROXY_PORT] [USER@]HOST[:PORT] [-k IDENTITY_FILE] [-p PORT]
```

|参数| 说明 | 是否必填| 默认值 |
|:---|:---|:---|:---|
|`USER`| 要登录的操作系统用户 | 否 | `root` |
|`HOST`| 目标服务器IP，目前仅支持 IPv4 | 是 ||
|`PORT`| 服务器 sshd 服务的端口| 否 | 22 |
|`IDENTITY_FILE`| 私钥文件 | 否 | `~/.ssh/id_rsa` |
|`-J`| 支持通过跳板机登录，跳板机的信息通过 -J 提供，跳板机目前仅支持密码登录 | 否 | |

当首次登录，不存在可用私钥时，会通过交互方式来让用户输入密码，一旦登录成功，这个密码就会被 ssx 保存到本地的数据文件中 (默认为 **~/.ssx/db**， 可通过环境变量 `SSX_DB_PATH` 进行自定义)。

下次登录时，直接执行 `ssx <IP>` 即可自动登录。

同时，为了简洁也可通过 `<IP>` 的片段直接搜索匹配登录，比如存储了一个条目 `192.168.1.100`，那么可以直接通过 `ssx 100` 即可登录。

## 查看条目列表

```bash
ssx list
# output example
# Entries (stored in ssx)
#  ID |       Address        |          Tags
#-----+----------------------+--------------------------
#  1  | root@172.23.1.84:22  | centos
```

ssx 默认不加载 `~/ssh/config` 文件，除非设置了环境变量 `SSX_IMPORT_SSH_CONFIG`。

ssx 不会将用户的 ssh 配置文件中的条目存储到自己的数据库中，因此您不会在 list 命令的输出中看到 “ID” 字段。

```bash
export SSX_IMPORT_SSH_CONFIG=true
ssx list
# output example
# Entries (stored in ssx)
#  ID |       Address        |          Tags
#-----+----------------------+--------------------------
#  1  | root@172.23.1.84:22  | centos
#
# Entries (found in ssh config)
#               Address              |           Tags
# -----------------------------------+----------------------------
#   git@ssh.github.com:22            | github.com
```

## 为条目打标签

ssx 会给每个存储的服务器分配一个唯一的 `ID`，我们在打标签时就需要通过 `ID` 来指定服务器条目。

打标签需要通过 ssx 的 `tag` 子命令来完成，下面是 tag 命令的模式：

```bash
ssx tag -i <ENTRY_ID> [-t TAG1 [-t TAG2 ...]] [-d TAG3 [-d TAG4 ...]]
```

- `-i`: 指定 list 命令输出的要操作的服务器对应的 ID 字段
- `-t`: 指定要添加的标签名，可以多次指定就可以同时添加多个标签
- `-d`: 打标签的同时也支持删除已有标签，通过 -d 指定要删除的标签名，同样也可以多次指定

当我们完成对服务器的打标签后，比如假设增加了一个 `centos` 的标签，那么我此时就可以通过标签来进行登录了：

```bash
ssx centos
```

## 登录服务器

如果没有指定任何参数标志，ssx 将把第二个参数作为搜索关键词，从主机和标签中搜索，如果没有匹配任何条目，ssx将把它作为一个新条目，并尝试登录。

```bash
# 通过交互登录，只需运行SSX
ssx

# 按条目id登录
ssx -i <ID>

# 通过地址登录，支持部分单词
ssx <ADDRESS>

# 通过标签登录
ssx <TAG>
```

## 执行单次命令

SSX 支持通过 `-c` 参数指定一个 shell 命令，登录后执行该命令后退出，便于一些嵌入式场景非交互的方式执行远程命令

```bash
ssx centos -c 'pwd'
```

## 升级SSX

> v0.3.0+

```bash
ssx upgrade [<version>]
```

默认不指定版本时会自动更新到 github 上的最新版。
