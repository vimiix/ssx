# 发布日志

## v0.5.0

发布时间：2024年11月14日

### BREAKING CHANGE

- Changed the SSH identity file flag from `-k` to `-i` to be more compatible with the standard `ssh` command
- Changed the entry ID flag from `-i` to `--id` across all commands for consistency
- Updated command examples and help text to reflect the new flag names

### Why

- The `-i` flag is the standard flag for specifying identity files in SSH, making SSX more intuitive for users familiar with SSH
- Using `--id` for entry IDs makes the parameter name more descriptive and avoids conflict with the SSH identity file flag

### Example Usage

```text
Old: ssx delete -i 123
New: ssx delete --id 123

Old: ssx -k ~/.ssh/id_rsa
New: ssx -i ~/.ssh/id_rsa
```

## v0.4.3

发布时间：2024年9月20日

**Bug Fix:**

- 修复Mac m1中使用时，会出现unexpected fault address 0xxxxx的问题 ([#62](https://github.com/vimiix/ssx/issues/62))

## v0.4.2

发布时间：2024年9月18日

**Changelog:**

- 更新依赖库版本

## v0.4.1

发布时间：2024年8月28日

**Changelog:**

- 更新依赖库版本

## v0.4.0

发布时间：2024年7月10日

**Features:**

- 强制要求给数据库文件设置管理员密码，未设置首次登录会要求补充
- 新增环境变量 `SSX_DEVIVE_ID`, 默认采用设备ID，数据库文件会绑定设备，如果迁移到其他机器需校验管理员密码后更新设备ID

**BREAKING CHANGE:**

- 废弃参数 `--unsafe`
- 废弃环境变量 `SSX_UNSAFE_MODE` 和 `SSX_SECRET_KEY`
- 如果旧版本存在 safe 模式的条目，需要在登录时重新输入一次密码

## v0.3.1

发布时间：2024年6月18日

**Features:**

- 升级时校验当前版本和最新版，避免重复升级
- 升级时无需加载条目数据库，减少非必要逻辑

## v0.3.0

发布时间：2024年6月12日

**Features:**

- 支持通过 `-k` 参数刷新存储的密钥记录
- 支持在线升级

**BREAKING CHANGE:**

- 将 `--server` 和 `--tag` 标记为已弃用参数

## v0.2.0

发布时间：2024年6月11日

**Features:**

- 新增参数 `-p` 支持显式指定端口
- 新增参数 `-J` 支持通过跳板机登录
- 默认使用设备ID加密条目密码
- 默认登录用户由当前用户修改为 root

## v0.1.0

发布时间：2024年2月29日

**Features:**

- 完成初版设计的预期需求，实现最小可用版本。
