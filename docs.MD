### Create minc-service -> 3
```
sudo docker service create \
--mode replicated \
--replicas 3 \
--name minc-service \
--network swarm-network \
--endpoint-mode dnsrr \
--limit-memory 100m \
tanmoysrt/minc:v2
```

### haproxy.cfg
```
global
  chroot /var/lib/haproxy
  master-worker
  stats socket /var/run/haproxy.sock user haproxy group haproxy mode 660 level admin expose-fd listeners

defaults unnamed_defaults_1
  mode http
  maxconn 3000
  log global
  timeout http-request 10s
  timeout check 10s
  timeout connect 10s
  timeout client 1m
  timeout queue 1m
  timeout server 1m
  timeout http-keep-alive 10s
  retries 3

userlist haproxy-dataplaneapi
  user admin insecure-password mypassword

frontend main
  bind *:80
  use_backend stat if { path -i /stats }
  default_backend error_backend

backend error_backend
  mode http
  http-request deny deny_status 400

backend stat
  stats enable
  stats uri /stats
  stats refresh 15s
  stats show-legends
  stats show-node

program api
  command /usr/local/bin/dataplaneapi --host 0.0.0.0 --port 5555 --haproxy-bin /usr/local/sbin/haproxy --config-file /usr/local/etc/haproxy/haproxy.cfg --userlist haproxy-dataplaneapi --reload-cmd "service haproxy reload" --restart-cmd "service haproxy restart" --reload-delay 5
  no option start-on-reload
```
### Create Service
```
sudo docker service create \
  --mode global \
  --name haproxy-service \
  --network swarm-network \
  --publish published=80,target=80,protocol=tcp,mode=host \
  --publish published=443,target=443,protocol=tcp,mode=host \
  --publish published=5555,target=5555,protocol=tcp,mode=host \
  --mount type=bind,src=/etc/haproxy,dst=/usr/local/etc/haproxy \
  --dns=127.0.0.11 \
  haproxytech/haproxy-debian:2.9
```

### Register new backend
```
Pre-requesties
> Add resolver and a dns configuration to it [docker]
----------------
1. Add backend
2. Add server template to backend
3. Add ACL to frontend
4. Add backend switching for created ACL

Without SSL, also possible

use_backend be_minc_service if { hdr(host) -i apache.tanmoy.info }
```

### Might need to enable support for letsencrypt
```
bind :443 ssl crt /etc/haproxy/certs/ alpn h2,http/1.1

redirect scheme https code 301 if !{ ssl_fc }

use_backend letsencrypt-backend if { path_beg /.well-known/acme-challenge/ }
```

> We should use multiple frontend rather than too much complex ssl

### Write program to issue SSL and upload to haproxy
```
- Try to have a golang program which will verify DNS records and issue certificates
- Also run SSL reloader every day to issue certificates before 30 days expiry
- Allow to import custom SSL certifcate
- Update in haproxy via API. provide details via volume mount
```