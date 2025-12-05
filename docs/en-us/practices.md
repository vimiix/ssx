# Practices

## Add hostname as tag for all machines

```bash
ssx list | grep -E '\s\d' | awk '{print $1}' | xargs -I id sh -c 'ssx --id id -c hostname|xargs ssx tag --id id -t '
```

## View server information

```bash
ssx info --id <ID>
```

## Get server IP

```bash
ssx info --id <ID> | jq .host
```

## Login via jump server

```bash
# Single jump server
ssx -J [proxy_user@]proxy_host[:proxy_port] [user@]host[:port]

# Multiple jump servers
ssx -J address1,address2,...  target_address
```

## Batch backup remote files to local

```bash
# Download config files from multiple servers
for tag in server1 server2 server3; do
  ssx cp $tag:/etc/nginx/nginx.conf ./backup/${tag}_nginx.conf
done
```

## Sync files between two servers

```bash
# Copy logs from production to backup server
ssx cp prod:/var/log/app.log backup:/var/log/app.log
```

## Quick upload deployment files using tags

```bash
# Upload deployment package to all web servers
ssx cp ./deploy.tar.gz webserver:/opt/deploy/
```
