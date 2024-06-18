
### 服务
`/usr/lib/systemd/system/podman-compose-daemon.service`
```shell
cat > /usr/lib/systemd/system/podman-compose-daemon.service <<EOF  
[Unit]
Description=podman-compose-daemon
Requires=podman.socket
After=podman.socket


[Service]
Type=simple 
ExecStart=podman-compose startup


[Install]
WantedBy = multi-user.target
EOF
```