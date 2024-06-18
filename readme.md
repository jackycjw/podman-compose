### 可执行文件
将`podman-compose-amd64`或者`podman-compose-arm64`放到 
`/usr/bin/podman-compose`或者`/usr/local/bin/podman-compose`

### 系统服务
执行下面的脚本
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
### 开机启动
```shell
systemctl enable --now podman-compose-daemon.service
```