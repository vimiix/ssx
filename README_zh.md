<p align="center">
    <img src="https://raw.githubusercontent.com/vimiix/ssx/master/static/logo.svg?sanitize=true"
        height="130">
</p>

<p align="center">
    <a href="https://github.com/vimiix/ssx/actions" alt="license">
    <img src="https://github.com/vimiix/ssx/actions/workflows/release.yml/badge.svg" /></a>
    <a href="https://goreportcard.com/report/github.com/vimiix/ssx" alt="goreport">
    <img src="https://goreportcard.com/badge/github.com/vimiix/ssx" /></a>
    <a href="https://github.com/vimiix/ssx/blob/main/LICENSE" alt="license">
    <img src="https://img.shields.io/badge/License-MIT-jasper" /></a>
    <a href="https://github.com/vimiix" alt="author">
    <img src="https://img.shields.io/badge/Author-Vimiix-blue" /></a>
</p>

<p align="center"><a href="https://github.com/vimiix/ssx/blob/main/README.md">English</a> | <a href="https://github.com/vimiix/ssx/blob/main/README_zh.md">中文</a></p>


🦅 SSX 是一个有记忆的 SSH 客户端。

它会自动记住通过它登录的服务器，因此，当您再次登录时，无需再次输入密码。

<p align="center">
    <img src="https://raw.githubusercontent.com/vimiix/ssx/master/static/demo.svg?sanitize=true"
        height="500">
</p>

## 需求来源

对于一个后端程序员来说，在工作中免不了要和繁杂的服务器打交道，ssh 是不可或缺的开发工具。但每次登录都需要输入密码的行为，对于认为一切皆可自动化的程序员来说，肯定是有点繁琐的（如果您是使用图形化界面的用户可忽略）。

所以我在前段时间考虑，我应该自己实现一个 ssh 客户端，它不需要拥有许多复杂的功能，只需要满足我以下这几个需求即可满足日常使用：

- 和 ssh 保持差不多的使用习惯
- 仅在第一次登录时询问我密码，后续使用无需再提供密码
- 可以给服务器它任意的标签，这样我就可以自由地通过IP 或者标签来登录

于是乎，近期我在业余时间就设计并编写了 ssx 这个轻量级的具有记忆的 ssh 客户端。它完美的实现了上面我所需要的功能，也已经被我愉快的应用到了日常的开发中，大大提高了搬砖效率。

## 使用方式

### 安装

ssx 是通过 golang 开发的一个独立的二进制文件，安装方式就是从 release 页面下载对应平台的软件包，解压后把 `ssx` 二进制放到系统的任意目录下，这里我习惯放到 `/usr/local/bin` 目录下，如果你选择其他目录下，需要确保存放的目录添加到 `$PATH` 环境变量中，这样后续使用我们就不用再添加路径前缀，直接通过 `ssx` 命令就可以运行了。

如果你想从源代码安装，你可以在项目根目录下运行命令:

```bash
make ssx
```

然后你可以在 **dist** 目录下得到 ssx 的二进制文件。

### 添加新条目(登录一次即代表新增)

```bash
ssx [USER@]HOST[:PORT] [-k IDENTITY_FILE] [-p PORT]
```

> 如果给定的地址与一个存在的条目匹配，ssx 将直接登录。

在这个命令中，`USER` 是可以省略的，如果省略则是系统当前用户名；`PORT` 是可以省略的，默认是 `22`，`-k IDENTITY_FILE` 代表的是使用私钥登录，通过 `-k` 来指定私钥的路径，也是可以省略的，默认是 `~/.ssh/id_rsa`，当然了，前提是这个文件存在。所以精简后的登录命令就是：`ssx <ip>`

当首次登录，不存在可用私钥时，会通过交互方式来让用户输入密码，一旦登录成功，这个密码就会被 ssx 保存到本地的数据文件中 (默认为 **~/.ssx/db**， 可通过环境变量 `SSX_DB_PATH` 进行自定义)，下次登录时，仍然执行 `ssx <ip>` 即可自动登录。

注意，登录过的服务器，再次登录时，我嫌输入全部 IP 比较繁琐，所以 ssx 支持输入 IP 中的部分字符，自动搜索匹配进行登录。

### 列出存在的条目

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

### 为服务器打标签

ssx 会给每个存储的服务器分配一个唯一的 `ID`，我们在打标签时就需要通过 `ID` 来指定服务器条目。

打标签需要通过 ssx 的 `tag` 子命令来完成，下面是 tag 命令的模式：

```bash
ssx tag -i <ENTRY_ID> [-t TAG1 [-t TAG2 ...]] [-d TAG3 [-d TAG4 ...]]
```

- -i 指定 list 命令输出的要操作的服务器对应的 ID 字段
- -t 指定要添加的标签名，可以多次指定就可以同时添加多个标签
- -d 打标签的同时也支持删除已有标签，通过 -d 指定要删除的标签名，同样也可以多次指定

当我们完成对服务器的打标签后，比如假设增加了一个 `centos` 的标签，那么我此时就可以通过标签来进行登录了：

```bash
ssx centos
```

### 登录服务器

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

### 执行命令

类似 ssh，ssx 也支持非交互式地执行指定的 shell 命令，可通过 `-c` 参数执行单条命令，如果没有执行 -c, ssx 会将第二个参数及其后面的所有参数均视为 shell 命令

```bash
ssx <ADDRESS> [-c] <COMMAND> [--timeout 30s]
ssx <TAG> [-c] <COMMAND> [--timeout 30s]

# 例如:登录192.168.1.100，执行命令'pwd':
ssx 1.100 pwd
# 通过 centos 标签执行
ssx centos [-c] pwd
```

### 删除服务器条目

```bash
ssx delete -i <ENTRY_ID>
```

## 支持的环境变量

- `SSX_DB_PATH`: 用于存储条目的数据库文件，默认为 **~/.ssx.db**；
- `SSX_CONNECT_TIMEOUT`: SSH连接超时，默认为: `10s`；
- `SSX_IMPORT_SSH_CONFIG`: 是否导入用户ssh配置，默认为空。
- `SSX_UNSAFE_MODE`: 密码以不安全模式存储
- `SSX_SECRET_KEY`: 用于加密条目密码的密钥，默认使用所在服务器的设备ID

这里解释一下 `SSX_IMPORT_SSH_CONFIG` 的作用，这个环境变量不设置时，ssx 默认是不会读取用户的 `~/.ssh/config` 文件的，ssx 只使用自己存储文件进行检索。如果将这个环境变量设置为非空（任意字符串），ssx 就会在初始化的时候加载用户 ssh 配置文件中存在的服务器条目，但 ssx 仅读取用于检索和登录，并不会将这些条目持久化到 ssx 的存储文件中，所以，如果 `ssx IP` 登录时，这个 `IP` 是 `~/.ssh/config` 文件中已经配置过登录验证方式的服务器，ssx 匹配到就直接登录了。但 ssx list 查看时，该服务器会被显示到 `found in ssh config` 的表格中，这个表格中的条目是不具有 ID 属性的。

## Upgrade SSX

> 新增于: v0.3.0

```bash
ssx upgrade
```

## 版权

© 2023-2024 Vimiix

在 MIT 许可协议下分发。可查看 [LICENSE](https://github.com/vimiix/ssx/blob/main/LICENSE) 文件详情
