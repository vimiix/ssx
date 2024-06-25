# 实践

## 为所有的机器添加 hostname 的标签

```bash
ssx list | grep -E '\s\d' | awk '{print $1}' | xargs -I id sh -c 'ssx -i id -c hostname|xargs ssx tag -i id -t '
```

## 查看某个服务器的信息

```bash
ssx info -i <ID>
```

## 获取服务器的IP

```bash
ssx info -i <ID> | jq .host
```

## 使用跳板机登录目标服务器

```bash
# 单个跳板机
ssx -J [proxy_user@]proxy_host[:proxy_port] [user@]host[:port]

# 多层跳板机
ssx -J address1,address2,...  target_address
```
