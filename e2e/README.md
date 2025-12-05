# SSX E2E 测试

本目录包含 ssx 命令行工具的端到端（E2E）测试用例集。

## 环境变量配置

运行测试前需要设置以下环境变量：

| 变量名 | 必需 | 说明 | 默认值 |
|--------|------|------|--------|
| `SSX_E2E_HOST` | 是* | SSH 服务器主机名或 IP | - |
| `SSX_E2E_PORT` | 否 | SSH 服务器端口 | 22 |
| `SSX_E2E_USER` | 是* | SSH 用户名 | - |
| `SSX_E2E_PASSWORD` | 是** | SSH 密码 | - |
| `SSX_E2E_KEY` | 是** | SSH 私钥文件路径 | - |
| `SSX_E2E_HOST2` | 否 | 第二台 SSH 服务器（用于 remote-to-remote 测试） | - |
| `SSX_E2E_PORT2` | 否 | 第二台 SSH 服务器端口 | 22 |
| `SSX_E2E_USER2` | 否 | 第二台 SSH 服务器用户名（默认使用 SSX_E2E_USER） | - |

> \* 需要 SSH 服务器的测试用例必需  
> \*\* `SSX_E2E_PASSWORD` 和 `SSX_E2E_KEY` 至少需要设置一个

## 运行测试

### 运行不需要服务器的测试

```bash
go test -v ./e2e/... -run "TestVersion|TestHelp|TestCpLocalToLocal|TestCpMissingArgs"
```

### 运行完整测试

```bash
SSX_E2E_HOST=192.168.1.100 \
SSX_E2E_USER=root \
SSX_E2E_PASSWORD=your_password \
go test -v ./e2e/...
```

### 使用 SSH 密钥认证

```bash
SSX_E2E_HOST=192.168.1.100 \
SSX_E2E_USER=root \
SSX_E2E_KEY=~/.ssh/id_rsa \
go test -v ./e2e/...
```

### 运行 remote-to-remote 测试

```bash
SSX_E2E_HOST=192.168.1.100 \
SSX_E2E_USER=root \
SSX_E2E_PASSWORD=your_password \
SSX_E2E_HOST2=192.168.1.101 \
go test -v ./e2e/... -run "TestCpRemoteToRemote"
```

## 测试用例说明

### version_test.go - 版本和帮助信息

| 测试用例 | 说明 | 需要服务器 |
|----------|------|:----------:|
| `TestVersion` | 测试 `--version` 输出 | 否 |
| `TestHelp` | 测试 `--help` 输出 | 否 |
| `TestCpHelp` | 测试 `cp --help` 输出 | 否 |

### list_test.go - 列表功能

| 测试用例 | 说明 | 需要服务器 |
|----------|------|:----------:|
| `TestListEmpty` | 测试空数据库时的列表输出 | 否 |
| `TestListAliases` | 测试 `l`/`ls` 别名 | 否 |
| `TestListAfterConnection` | 测试连接后的列表显示 | 是 |

### connect_test.go - 连接功能

| 测试用例 | 说明 | 需要服务器 |
|----------|------|:----------:|
| `TestConnectAndExecute` | 测试连接并执行命令 | 是 |
| `TestConnectWithPort` | 测试指定端口连接 | 是 |
| `TestConnectWithIdentityFile` | 测试使用 SSH 密钥连接 | 是 |
| `TestConnectByKeyword` | 测试通过关键字匹配连接 | 是 |
| `TestConnectWithTimeout` | 测试命令超时功能 | 是 |
| `TestConnectTimeoutExceeded` | 测试超时中断 | 是 |

### tag_test.go - 标签功能

| 测试用例 | 说明 | 需要服务器 |
|----------|------|:----------:|
| `TestTagAddAndDelete` | 测试添加和删除标签 | 是 |
| `TestTagRequiresID` | 测试缺少 `--id` 参数时的错误 | 否 |
| `TestTagNoTagSpecified` | 测试未指定标签时的错误 | 否 |
| `TestConnectByTag` | 测试通过标签连接服务器 | 是 |

### delete_test.go - 删除功能

| 测试用例 | 说明 | 需要服务器 |
|----------|------|:----------:|
| `TestDeleteEntry` | 测试删除单个条目 | 是 |
| `TestDeleteMultipleEntries` | 测试批量删除条目 | 是 |
| `TestDeleteNoID` | 测试未指定 ID 时的错误 | 否 |
| `TestDeleteAliases` | 测试 `d`/`del` 别名 | 否 |

### info_test.go - 信息查询功能

| 测试用例 | 说明 | 需要服务器 |
|----------|------|:----------:|
| `TestInfoByID` | 测试通过 ID 查询条目信息 | 是 |
| `TestInfoByKeyword` | 测试通过关键字查询条目信息 | 是 |
| `TestInfoByTag` | 测试通过标签查询条目信息 | 是 |
| `TestInfoNotFound` | 测试查询不存在条目时的错误 | 否 |
| `TestInfoPasswordMasked` | 测试密码在输出中被掩码 | 是 |

### cp_test.go - 文件复制功能

| 测试用例 | 说明 | 需要服务器 |
|----------|------|:----------:|
| `TestCpUpload` | 测试上传本地文件到远程 | 是 |
| `TestCpDownload` | 测试下载远程文件到本地 | 是 |
| `TestCpWithTag` | 测试使用标签引用远程主机 | 是 |
| `TestCpRemoteToRemote` | 测试远程到远程文件复制 | 是（双服务器） |
| `TestCpLocalToLocal` | 测试本地到本地复制被拒绝 | 否 |
| `TestCpMissingArgs` | 测试缺少参数时的错误 | 否 |
| `TestCpNonExistentLocalFile` | 测试上传不存在的文件时的错误 | 是 |

## 测试文件结构

```
e2e/
├── README.md           # 本文件
├── e2e_test.go         # 测试框架和辅助函数
├── version_test.go     # 版本和帮助测试
├── list_test.go        # 列表功能测试
├── connect_test.go     # 连接功能测试
├── tag_test.go         # 标签功能测试
├── delete_test.go      # 删除功能测试
├── info_test.go        # 信息查询测试
└── cp_test.go          # 文件复制测试
```

## 注意事项

1. **测试会自动编译 ssx 二进制文件**：测试开始时会在临时目录编译最新的 ssx 二进制文件
2. **测试使用独立数据库**：每个测试使用独立的临时数据库，不会影响本地 ssx 配置
3. **测试会在远程服务器创建临时文件**：文件复制测试会在 `/tmp` 目录创建临时文件，测试结束后会自动清理
4. **跳过机制**：缺少必要环境变量的测试会被自动跳过，不会报错
