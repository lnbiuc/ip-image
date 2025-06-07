![image.png](https://cdn.violet.vin/v2/S8CVLHD.png)

# 介绍
获取客户端IP地址、AS信息，返回一张png图片，用于teamspeak服务器bunner展示。

基于ip-api（限制qps45）和百度ip定位api

# Build

```ps1
$env:GOOS = "linux"
$env:GOARCH = "amd64"
go build -o go-ip-info-linux-amd64 main.go
```

```cmd
set GOOS=linux
set GOARCH=amd64
go build -o go-ip-info-linux-amd64 main.go
```
# 使用

```service
[Unit]
Description=Go IP Info Web Service
After=network.target

[Service]
WorkingDirectory=/root/ipinfo
ExecStart=/root/ipinfo/go-ip-info-linux-amd64
Environment=BAIDU_API_KEY=<替换成百度APIKEY>
Restart=always
RestartSec=3
StandardOutput=append:/root/ipinfo/output.log
StandardError=append:/root/ipinfo/error.log

[Install]
WantedBy=multi-user.target
```

```bash
sudo systemctl daemon-reload
sudo systemctl enable go-ip-info.service
sudo systemctl start go-ip-info.service
sudo systemctl status go-ip-info.service
```